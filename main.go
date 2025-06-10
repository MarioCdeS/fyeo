package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/MarioCdeS/fyeo/internal/config"
	"github.com/MarioCdeS/fyeo/internal/secretservice"
)

const (
	defaultCfgFile = "fyeo.config.yml"
	secretTimeout  = 3 * time.Second
)

var (
	cfgFile  *string
	showHelp *bool
)

func init() {
	pflag.CommandLine.SetInterspersed(false)
	pflag.Usage = printUsage

	cfgFile = pflag.String("config", defaultCfgFile, "Configuration file path")
	showHelp = pflag.BoolP("help", "h", false, "Show help message")
}

func main() {
	pflag.Parse()

	if *showHelp {
		printUsage()
		return
	}

	args := pflag.Args()

	if len(args) == 0 {
		printUsage()
		return
	}

	exe := must(exec.LookPath, args[0], "Unable to launch the target executable")
	cfg := must(config.LoadFromFile, *cfgFile, "Unable to load the configuration")
	secrets := must(fetchSecrets, cfg, "Unable to retrieve secret values")
	env := buildEnvironment(cfg, secrets)

	if err := syscall.Exec(exe, args, env); err != nil {
		printError("Unable to launch target command", err)
		os.Exit(1)
	}
}

func fetchSecrets(cfg *config.Config) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), secretTimeout)
	defer cancel()

	service, err := secretservice.New(cfg.ProjectID, ctx)

	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(cfg.Secrets))

	for _, s := range cfg.Secrets {
		v, err := func() (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), secretTimeout)
			defer cancel()
			return service.Fetch(s.Name, s.Version, ctx)
		}()

		if err != nil {
			return nil, err
		}

		values[s.Name] = v
	}

	return values, nil
}

func buildEnvironment(cfg *config.Config, secrets map[string]string) []string {
	env := make([]string, len(cfg.Secrets))

	for i, s := range cfg.Secrets {
		if val, ok := secrets[s.Name]; ok {
			env[i] = fmt.Sprintf("%s=%s", s.Env, val)
		} else {
			panic(fmt.Sprintf("unexpected error - no secret value found for %s", s.Name))
		}
	}

	return append(os.Environ(), env...)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [--config <config-file>] [-h | --help] [--] COMMAND\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "COMMAND: Target command to execute")
	fmt.Fprintln(os.Stderr, "Flags:")
	pflag.PrintDefaults()
}

func printError(msg string, cause error) {
	fmt.Fprintln(os.Stderr, "Error:", msg)

	if cause != nil {
		fmt.Fprintf(os.Stderr, "Cause: %+v\n", cause)
	}
}

func must[A, R any](f func(A) (R, error), arg A, msg string) R {
	res, err := f(arg)

	if err != nil {
		printError(msg, err)
		os.Exit(1)
	}

	return res
}

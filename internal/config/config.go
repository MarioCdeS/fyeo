package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"

	"github.com/MarioCdeS/fyeo/internal/errors"
)

type Config struct {
	ProjectID string `mapstructure:"project-id"`
	Secrets   []struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
		Env     string `mapstructure:"env"`
	} `mapstructure:"secrets"`
}

func LoadFromFile(file string) (*Config, error) {
	viper.SetConfigFile(file)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.NewWithCause("failed to read the configuration file", err)
	}

	var cfg Config

	if err := viper.UnmarshalExact(&cfg); err != nil {
		return nil, errors.NewWithCause("failed to parse the configuration file", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, errors.NewWithCause("invalid configuration settings", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	var errs []error

	if cfg.ProjectID == "" {
		errs = append(errs, errors.New("project-id may not be an empty string"))
	}

	for i, s := range cfg.Secrets {
		if s.Name == "" {
			errs = append(errs, errors.New(fmt.Sprintf("secrets[%d].name may not be an empty string", i)))
		}

		if s.Version == "" {
			errs = append(errs, errors.New(fmt.Sprintf("secrets[%d].version may not be an empty string", i)))
		} else if s.Version != "latest" {
			if v, err := strconv.Atoi(s.Version); err != nil || v < 1 {
				msg := fmt.Sprintf(
					"secrets[%d].version must be valid version number or alias (\"%s\" is invalid)",
					i,
					s.Version,
				)
				errs = append(errs, errors.New(msg))
			}
		}

		if s.Env == "" {
			errs = append(errs, errors.New(fmt.Sprintf("secrets[%d].env may not be an empty string", i)))
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

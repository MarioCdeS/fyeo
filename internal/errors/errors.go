package errors

import (
	goerrs "errors"
	"fmt"
	"iter"
	"strings"
)

type Error struct {
	msg   string
	cause error
}

func New(msg string) Error {
	return NewWithCause(msg, nil)
}

func NewWithCause(msg string, cause error) Error {
	return Error{msg, cause}
}

func (e Error) Error() string {
	if e.cause == nil {
		return e.msg
	} else {
		return fmt.Sprintf("%s: %s", e.msg, e.cause)
	}
}

func (e Error) Unwrap() error {
	return e.cause
}

func (e Error) Format(f fmt.State, verb rune) {
	if verb != 'v' || !f.Flag('+') || e.cause == nil {
		fmt.Fprint(f, e.Error())
		return
	}

	var sb strings.Builder
	sb.WriteString(e.msg)

	causeLines := strings.Lines(fmt.Sprintf("%+v", e.cause))
	next, stop := iter.Pull(causeLines)
	defer stop()

	if first, ok := next(); ok {
		sb.WriteString("\n └─ ")
		sb.WriteString(first)
	}

	for line, ok := next(); ok; line, ok = next() {
		sb.WriteString("    ")
		sb.WriteString(line)
	}

	fmt.Fprint(f, sb.String())
}

func Join(errs ...error) error {
	return goerrs.Join(errs...)
}

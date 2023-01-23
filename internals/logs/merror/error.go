package merror

import (
	"fmt"

	"github.com/pkg/errors"
)

const errFmt = "ERROR, %s"

func Error(msg string) error {
	return fmt.Errorf(errFmt, msg)
}

func Errorf(format string, a ...interface{}) error {
	return Error(fmt.Sprintf(format, a...))
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return errors.Wrap(err, fmt.Sprintf(errFmt, msg))
}

func Wrapf(err error, format string, a ...interface{}) error {
	return Wrap(err, fmt.Sprintf(format, a...))
}

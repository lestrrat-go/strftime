//go:build strftime_native_errors

package errors

import "fmt"

func New(s string) error {
	return fmt.Errorf(s)
}

func Wrap(err error, s string) error {
	return fmt.Errorf(s+`: %w`, err)
}

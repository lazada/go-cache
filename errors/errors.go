package errors

import "github.com/pkg/errors"

var (
	New    func(msg string) error
	Wrap   func(err error, message string) error
	Wrapf  func(err error, format string, args ...interface{}) error
	Errorf func(format string, args ...interface{}) error
)

func init() {
	SetWithStack(false)
}

func SetWithStack(errorTraces bool) {
	New = errors.New
	Wrap = errors.Wrap
	Wrapf = errors.Wrapf
	Errorf = errors.Errorf

	if errorTraces {
		New = func(msg string) error {
			return errors.WithStack(New(msg))
		}

		Wrap = func(err error, message string) error {
			return errors.WithStack(Wrap(err, message))
		}

		Wrapf = func(err error, format string, args ...interface{}) error {
			return errors.WithStack(Wrapf(err, format, args...))
		}

		Errorf = func(format string, args ...interface{}) error {
			return errors.WithStack(Errorf(format, args...))
		}
	}
}

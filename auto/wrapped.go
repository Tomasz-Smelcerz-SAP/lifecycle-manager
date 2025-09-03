package auto

import (
	"errors"
	"fmt"
	"runtime"
)

const (
	errFormat      = "[at %s:%d (%s)]: %%s"
	maxFileNameLen = 20
)

var ErrWrapped = errors.New("sentinel")

type WrappedError struct {
	msgTemplate string
	err         error
}

func (e *WrappedError) Is(target error) bool {
	return target == ErrWrapped //nolint:err113 // we can't use errors.Is, because we're CALLED by it.
}

func (e *WrappedError) Unwrap() error {
	return e.err
}

func (e *WrappedError) String() string {
	endIdx := len(e.msgTemplate) - 4 //nolint:mnd // the trailing part of the message has length of 4, I don't think using a const here improves readability.
	return e.msgTemplate[0:endIdx]
}

func (e *WrappedError) Error() string {
	return fmt.Sprintf(e.msgTemplate, e.err.Error())
}

func Wrap(err error) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("No caller information")
	}
	function := runtime.FuncForPC(pc)

	if len(file) > maxFileNameLen {
		file = "..." + file[len(file)-20:]
	}
	return &WrappedError{
		msgTemplate: fmt.Sprintf(errFormat, file, line, function.Name()),
		err:         err,
	}
}

This module introduces the "auto.Wrap()" function,
which automatically wraps errors in your Go code.

So instead of writing:

	if err != nil {
		return fmt.Errorf("failed to do something: %w", err)
	}

You can simply write:

	if err != nil {
		return auto.Wrap(err)
	}

The main purpose for this error wrapper is to wrap errors you don't really care about,
but you have to propagate them up the call stack because you don't want to "swallow" them.


Details:

The "auto.Wrap()" function will automatically add context to the error message,
including the file name and line number where the "Wrap()" was invoked.

This is how it renders the error message:
	[at ...al/idontcare/main.go:43 (main.doBusinessLogic)]: invalid reason

The "[at ...)]: " part is added by "auto.Wrap()",
the "invalid reason" is the original error message.
Of course this may be chained, so one auto-wrapped error may wrap
another auto-wrapped error and so on.
This leads to a stack-trace like error message.

The advantage of using "auto.Wrap()" over "fmt.Errorf()" is:
- YOU MAKE IT CLEAR that the auto-wrapped error is the one you don't really care about,
  and you just propagate it up the call stack
- you don't have to manually invent error messages for errors you don't care about
- because of the above it reduces boilerplate code: 'auto.Wrap(err)' is usually shorter than
  'fmt.Errorf("failed to do something: %w", err))'
- the error message includes the file name and line number, which can be useful for debugging
- the message format allows for filtering out the auto-wrapped "traces" from the logs if needed,
  or to highlight them if required.


package a

import "github.com/pkg/errors"

func foo() error { return nil }

// Not formatted with gofmt on purpose: the call and the wrap are on the same line,
// findCallData takes only calls with "callData.line < wrapLine".
func A() error {
	if err := foo(); err != nil { return errors.Wrap(err, "bar") } // NOT FOUND: message "bar", called func foo

	return nil
}

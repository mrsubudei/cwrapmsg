package a

import (
	"fmt"

	"github.com/pkg/errors"
)

func foo() error { return nil }

// When the outer wrap is skipped (cyrillic message with -ignore-cyrillic, message is not a string literal,
// fmt.Errorf without %w), ast.Inspect returns false and the wraps inside its arguments are not visited.

func A() error {
	return errors.Wrap(errors.Wrap(foo(), "bar"), "ошибка") // NOT FOUND: inner wrap, message "bar", called func foo
}

func B(msg string) error {
	return errors.Wrap(errors.Wrap(foo(), "bar"), msg) // NOT FOUND: inner wrap, message "bar", called func foo
}

func C() error {
	return fmt.Errorf("baz: %v", errors.Wrap(foo(), "bar")) // NOT FOUND: inner wrap, message "bar", called func foo
}

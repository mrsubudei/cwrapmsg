package a

import "github.com/pkg/errors"

func foo() error { return nil }

func bar() int { return 0 }

// For "a, b := x(), y()" every call on the right side is recorded for err,
// the last one is bar, but err is assigned from foo.
func A() error {
	err, n := foo(), bar()
	_ = n

	return errors.Wrap(err, "bar") // NOT FOUND: message "bar", err from foo
}

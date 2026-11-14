package a

import "fmt"

func foo() error { return nil }

// getWIndex counts only "%" followed by a letter, so verbs with flags, width, precision
// or explicit index (%+v, %5d, %[1]v) are skipped and the index of %w is wrong.

func A(x int) error {
	return fmt.Errorf("bar %+v: %w", x, foo()) // NOT FOUND: message "bar", called func foo (%w is taken as x)
}

func B(n int) error {
	err := foo()

	return fmt.Errorf("bar %5d: %w", n, err) // NOT FOUND: message "bar", err from foo (%w is taken as n)
}

func C(x int) error {
	return fmt.Errorf("bar %[2]w: %[1]v", x, foo()) // NOT FOUND: message "bar", called func foo (%[2]w is not counted at all)
}

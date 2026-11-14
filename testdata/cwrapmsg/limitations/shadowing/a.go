package a

import "github.com/pkg/errors"

func load() (int, error) { return 0, nil }

func set() error { return nil }

// The wrapped err is from load, but the inner err from set (another variable with the same name)
// is the last call before the wrap, so the message is compared with set.
func A() error {
	_, err := load()
	if err := set(); err != nil {
		println(err)
	}

	if err != nil {
		return errors.Wrap(err, "set") // NOT FOUND: message "set", err from load
	}

	return nil
}

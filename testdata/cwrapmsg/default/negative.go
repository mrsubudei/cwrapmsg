package defaults

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Correct messages, nothing is reported.

func (s *Service) WrapCorrect(ctx context.Context, ids []int) error {
	// the full name
	_, err := s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "s.store.GetByIDs")
	}

	// a part of the name without "s."
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "store.GetByIDs")
	}

	// only the name of the method
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "GetByIDs")
	}

	// underscores in the message are ignored
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "store.Get_By_IDs")
	}

	// underscores in the name of the func are ignored
	err = load_items()
	if err != nil {
		return errors.Wrap(err, "load_items")
	}

	err = load_items()
	if err != nil {
		return errors.Wrap(err, "loaditems")
	}

	// only letters, digits and dots before the first other symbol are compared
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrapf(err, "GetByIDs: ids: %v", ids)
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrapf(err, "GetByIDs %v", ids)
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "GetByIDs-failed")
	}

	// the same call is wrapped several times
	_, err = getItemIDs()
	if err != nil {
		return errors.Wrap(err, "getItemIDs")
	}
	if err != nil {
		return errors.Wrap(err, "getItemIDs")
	}

	return nil
}

func (s *Service) WrapCalls(ctx context.Context) error {
	// a method of the result of a call, only the last name is compared
	err := s.readConn().Execx("query")
	if err != nil {
		return errors.Wrap(err, "Execx")
	}

	err = s.readConn().Execx("query")
	if err != nil {
		return errors.Wrap(err, "readConn.Execx")
	}

	_, err = getStore().GetByIDs(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "GetByIDs")
	}

	// a method of an element of a slice
	_, err = s.items[0].GetByIDs(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "items.GetByIDs")
	}

	// generic funcs
	_, err = generic[int](5)
	if err != nil {
		return errors.Wrap(err, "generic")
	}

	_, err = generic2[string, int]("key", 5)
	if err != nil {
		return errors.Wrap(err, "generic2")
	}

	// the error is a direct func call
	if s != nil {
		return errors.Wrap(getError(), "getError")
	}

	// the error is declared with var
	var err2 = getError()
	if err2 != nil {
		return errors.Wrap(err2, "getError")
	}

	return nil
}

func ErrorfCorrect() error {
	_, err := getItemIDs()
	if err != nil {
		return fmt.Errorf("getItemIDs: %w", err)
	}

	// other verbs before and after %w
	_, err = getItemIDs()
	if err != nil {
		return fmt.Errorf("getItemIDs: id: %d, %w", 5, err)
	}

	_, err = getItemIDs()
	if err != nil {
		return fmt.Errorf("getItemIDs: %w. id: %d", err, 6)
	}

	// the error is a direct func call
	return fmt.Errorf("getError: %w", getError())
}

func Closures() error {
	// the call and the wrap are in the same closure
	err := run(func() error {
		_, err := getItemIDs()
		if err != nil {
			return errors.Wrap(err, "getItemIDs")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "run")
	}

	// nested closures
	err = run(func() error {
		err := run(func() error {
			err := getError()

			return errors.Wrap(err, "getError")
		})

		return errors.Wrap(err, "run")
	})

	return errors.Wrap(err, "run")
}

var packageLevelClosure = func() error {
	err := getError()

	return errors.Wrap(err, "getError")
}

// Not checked.

func NotChecked(ctx context.Context, err error, msg string) error {
	// the error comes from a parameter: it is not known which func returned it
	if err != nil {
		return errors.Wrap(err, "anything")
	}

	// the error is returned in another closure than the wrap
	var err2 error
	_ = run(func() error {
		_, err2 = getItemIDs()

		return nil
	})
	if err2 != nil {
		return errors.Wrap(err2, "anything")
	}

	// the message is not a string literal
	err = getError()
	if err != nil {
		return errors.Wrap(err, msg)
	}

	// fmt.Errorf without %w
	if err != nil {
		return fmt.Errorf("anything: %v", err)
	}

	// fmt.Errorf with several %w
	return fmt.Errorf("anything: %w, %w", err, getError())
}

// Wraps at package level are not checked (and do not hang).

var errWrappedBefore = errors.Wrap(errStub, "anything")

var errErrorfBefore = fmt.Errorf("anything: %w", errStub)

func init() {
	_, _ = errWrappedBefore, errErrorfBefore
}

var errWrappedAfter = errors.Wrap(errStub, "anything")

var errErrorfAfter = fmt.Errorf("anything: %w", errStub)

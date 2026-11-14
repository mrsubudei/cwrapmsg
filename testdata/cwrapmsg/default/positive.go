package defaults

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Incorrect messages, each wrap is reported.

func (s *Service) WrapIncorrect(ctx context.Context, ids []int) error {
	// another name
	_, err := s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "wrong") // want `message in errors.Wrap does not match function or method name`
	}

	// with spaces, only "get" is compared
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "get by ids") // want `message in errors.Wrap does not match function or method name`
	}

	// extra symbols before the name
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, ",GetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, ".GetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "zGetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	// extra word after the name
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "GetByIDs.extra") // want `message in errors.Wrap does not match function or method name`
	}

	// extra word before the full name
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "extra.s.store.GetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	// should be store.GetByIDs
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "s.GetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	// the comparison is case-sensitive by default
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "getByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	err = load_items()
	if err != nil {
		return errors.Wrap(err, "loadItems") // want `message in errors.Wrap does not match function or method name`
	}

	// messages with cyrillic symbols are checked by default
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "ошибка") // want `message in errors.Wrap does not match function or method name`
	}

	// details after the wrong name
	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrapf(err, "GetByID: ids: %v", ids) // want `message in errors.Wrapf does not match function or method name`
	}

	// the message for another error: err1 is from getItemIDs
	_, err1 := getItemIDs()
	_, err2 := s.store.GetByIDs(ctx, ids)
	if err2 != nil {
		return errors.Wrap(err1, "GetByIDs") // want `message in errors.Wrap does not match function or method name`
	}

	return nil
}

func (s *Service) WrapCallsIncorrect(ctx context.Context) error {
	// a method of the result of a call, only the last name is compared
	err := s.readConn().Execx("query")
	if err != nil {
		return errors.Wrap(err, "readConn") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = generic[int](5)
	if err != nil {
		return errors.Wrap(err, "generic2") // want `message in errors.Wrap does not match function or method name`
	}

	// the error is a direct func call
	if s != nil {
		return errors.Wrap(getError(), "getErr") // want `message in errors.Wrap does not match function or method name`
	}

	// the error is declared with var
	var err2 = getError()
	if err2 != nil {
		return errors.Wrap(err2, "getErr") // want `message in errors.Wrap does not match function or method name`
	}

	return nil
}

func ErrorfIncorrect() error {
	_, err := getItemIDs()
	if err != nil {
		return fmt.Errorf("getItems: %w", err) // want `message in fmt.Errorf does not match function or method name`
	}

	_, err = getItemIDs()
	if err != nil {
		return fmt.Errorf("get item ids: id: %d, %w", 5, err) // want `message in fmt.Errorf does not match function or method name`
	}

	_, err = getItemIDs()
	if err != nil {
		return fmt.Errorf("getItems: %w. id: %d", err, 6) // want `message in fmt.Errorf does not match function or method name`
	}

	return fmt.Errorf("getErr: %w", getError()) // want `message in fmt.Errorf does not match function or method name`
}

func ClosuresIncorrect() error {
	err := run(func() error {
		_, err := getItemIDs()
		if err != nil {
			return errors.Wrap(err, "getItems") // want `message in errors.Wrap does not match function or method name`
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "runFunc") // want `message in errors.Wrap does not match function or method name`
	}

	err = run(func() error {
		err := run(func() error {
			err := getError()

			return errors.Wrap(err, "run") // want `message in errors.Wrap does not match function or method name`
		})

		return errors.Wrap(err, "getError") // want `message in errors.Wrap does not match function or method name`
	})

	return errors.Wrap(err, "getError") // want `message in errors.Wrap does not match function or method name`
}

var packageLevelClosureIncorrect = func() error {
	err := getError()

	return errors.Wrap(err, "getErr") // want `message in errors.Wrap does not match function or method name`
}

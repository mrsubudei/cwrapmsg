package withflags

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Correct messages with -case-insensitive and -ignore-cyrillic, nothing is reported.

func (s *Service) CaseInsensitive(ctx context.Context, ids []int) error {
	_, err := s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "getbyids")
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "STORE.GETBYIDS")
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrapf(err, "get_by_ids: %v", ids)
	}

	err = load_items()
	if err != nil {
		return errors.Wrap(err, "loadItems")
	}

	err = s.readConn().Execx("query")
	if err != nil {
		return errors.Wrap(err, "execx")
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("store.getByIDs: %w", err)
	}

	return errors.Wrap(getError(), "geterror")
}

func IgnoreCyrillic() error {
	err := getError()
	if err != nil {
		return errors.Wrap(err, "ошибка")
	}

	if err != nil {
		return errors.Wrapf(err, "ошибка: %d", 5)
	}

	if err != nil {
		return fmt.Errorf("getError: ошибка %w", err)
	}

	return errors.Wrap(getError(), "ошибка")
}

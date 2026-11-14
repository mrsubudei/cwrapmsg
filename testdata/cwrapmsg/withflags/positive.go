package withflags

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Incorrect messages with -case-insensitive and -ignore-cyrillic, each wrap is reported.

func (s *Service) Incorrect(ctx context.Context, ids []int) error {
	_, err := s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "getbyidsz") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrap(err, "get by ids") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return errors.Wrapf(err, "s.getbyids: %v", ids) // want `message in errors.Wrapf does not match function or method name`
	}

	err = s.readConn().Execx("query")
	if err != nil {
		return errors.Wrap(err, "readconn") // want `message in errors.Wrap does not match function or method name`
	}

	_, err = s.store.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("getbyid: %w", err) // want `message in fmt.Errorf does not match function or method name`
	}

	return errors.Wrap(getError(), "geterr") // want `message in errors.Wrap does not match function or method name`
}

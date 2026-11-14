package defaults

import (
	stdfmt "fmt"

	pkgerrors "github.com/pkg/errors"
)

// Aliases of github.com/pkg/errors and fmt.

func Alias() error {
	_, err := getItemIDs()
	if err != nil {
		return pkgerrors.Wrap(err, "getItemIDs")
	}

	_, err = getItemIDs()
	if err != nil {
		return stdfmt.Errorf("getItemIDs: %w", err)
	}

	_, err = getItemIDs()
	if err != nil {
		return pkgerrors.Wrapf(err, "getItems: %d", 5) // want `message in errors.Wrapf does not match function or method name`
	}

	return stdfmt.Errorf("getItems: %w", err) // want `message in fmt.Errorf does not match function or method name`
}

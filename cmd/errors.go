package cmd

import (
	"errors"
	"fmt"

	"github.com/azure/azd.extensions.drasi/internal/output"
)

var errNotYetImplemented = errors.New("not yet implemented")

func notImplemented(name string) error {
	return fmt.Errorf("%s: %s: %w", output.ERR_NOT_IMPLEMENTED, name, errNotYetImplemented)
}

package drasi

import (
	"context"
	"fmt"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

// DeleteComponent deletes the named component. Returns nil if not found.
func (c *Client) DeleteComponent(ctx context.Context, kind, id string) error {
	_, stderr, exitCode, err := c.runner.Run(ctx, "delete", kind, id)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		if strings.Contains(stderr, "not found") {
			return nil
		}
		return fmt.Errorf("%s: drasi delete %s %s: %s", output.ERR_DRASI_CLI_ERROR, kind, id, strings.TrimSpace(stderr))
	}
	return nil
}

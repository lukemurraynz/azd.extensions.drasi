package drasi

import (
	"context"
	"fmt"
	"strings"

	"github.com/azure/azd.extensions.drasi/internal/output"
)

// WaitOnline polls until the component reaches Online status or timeout.
func (c *Client) WaitOnline(ctx context.Context, kind, id string, timeoutSec int) error {
	for callCount := 0; ; {
		stdout, _, _, err := c.runner.Run(ctx, "wait", "--kind", kind, "--id", id)
		callCount++
		if err != nil {
			return err
		}
		if strings.Contains(stdout, "STATUS: Online") {
			return nil
		}
		if callCount >= timeoutSec {
			return fmt.Errorf("%s: %s/%s timed out waiting to go online", output.ERR_COMPONENT_TIMEOUT, kind, id)
		}
	}
}

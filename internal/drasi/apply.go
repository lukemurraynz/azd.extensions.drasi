package drasi

import "context"

// ApplyFile applies a component manifest from the given file path.
func (c *Client) ApplyFile(ctx context.Context, path string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return c.RunCommand(ctx, "apply", "-f", path)
}

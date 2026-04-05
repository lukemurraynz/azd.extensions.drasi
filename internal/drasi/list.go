package drasi

import (
	"context"
	"fmt"
	"strings"
)

// ComponentSummary is a row from `drasi list <kind>` output.
type ComponentSummary struct {
	ID     string
	Kind   string
	Status string
}

// ListComponents lists all components of the given kind.
func (c *Client) ListComponents(ctx context.Context, kind string) ([]ComponentSummary, error) {
	stdout, _, _, err := c.runner.Run(ctx, "list", kind)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(stdout, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "ID KIND STATUS" {
		return nil, fmt.Errorf("unexpected drasi list output: missing or malformed header")
	}

	result := []ComponentSummary{}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("unexpected drasi list output: malformed row %q", line)
		}
		result = append(result, ComponentSummary{
			ID:     fields[0],
			Kind:   fields[1],
			Status: fields[2],
		})
	}
	return result, nil
}

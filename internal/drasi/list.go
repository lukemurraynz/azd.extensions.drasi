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
	return c.listComponents(ctx, kind, "")
}

// ListComponentsInContext lists all components of the given kind against a specific kube context.
func (c *Client) ListComponentsInContext(ctx context.Context, kind, kubeContext string) ([]ComponentSummary, error) {
	return c.listComponents(ctx, kind, kubeContext)
}

func (c *Client) listComponents(ctx context.Context, kind, kubeContext string) ([]ComponentSummary, error) {
	args := []string{"list", kind}
	if strings.TrimSpace(kubeContext) != "" {
		args = append([]string{"--context", kubeContext}, args...)
	}

	stdout, _, _, err := c.runner.Run(ctx, args...)
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

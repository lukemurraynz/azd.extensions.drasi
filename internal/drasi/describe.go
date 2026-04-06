package drasi

import (
	"context"
	"fmt"
	"strings"
)

// ComponentDetail is the parsed output of `drasi describe <kind> <id>`.
type ComponentDetail struct {
	ID          string
	Kind        string
	Status      string
	ErrorReason string
}

// ComponentNotFoundError identifies a missing component.
type ComponentNotFoundError struct {
	Kind string
	ID   string
}

// Error returns the error message.
func (e *ComponentNotFoundError) Error() string {
	return fmt.Sprintf("component not found: %s %s", e.Kind, e.ID)
}

// ErrComponentNotFound is returned when the component does not exist.
var ErrComponentNotFound error = &ComponentNotFoundError{}

// DescribeComponent returns full detail about a specific component.
func (c *Client) DescribeComponent(ctx context.Context, kind, id string) (*ComponentDetail, error) {
	return c.describeComponent(ctx, kind, id, "")
}

// DescribeComponentInContext returns full detail about a specific component against a specific kube context.
func (c *Client) DescribeComponentInContext(ctx context.Context, kind, id, kubeContext string) (*ComponentDetail, error) {
	return c.describeComponent(ctx, kind, id, kubeContext)
}

func (c *Client) describeComponent(ctx context.Context, kind, id, kubeContext string) (*ComponentDetail, error) {
	args := []string{"describe", kind, id}
	argsWithoutContext := args
	if strings.TrimSpace(kubeContext) != "" {
		args = append([]string{"--context", kubeContext}, args...)
	}

	stdout, stderr, exitCode, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	if exitCode != 0 {
		if strings.TrimSpace(kubeContext) != "" && isUnsupportedContextFlagError(stderr) {
			stdout, stderr, exitCode, err = c.runner.Run(ctx, argsWithoutContext...)
			if err != nil {
				return nil, err
			}
			if exitCode != 0 {
				if strings.Contains(stderr, "not found") {
					return nil, &ComponentNotFoundError{Kind: kind, ID: id}
				}
				return nil, fmt.Errorf("drasi describe %s %s: %s", kind, id, strings.TrimSpace(stderr))
			}
		} else {
			if strings.Contains(stderr, "not found") {
				return nil, &ComponentNotFoundError{Kind: kind, ID: id}
			}
			return nil, fmt.Errorf("drasi describe %s %s: %s", kind, id, strings.TrimSpace(stderr))
		}
	}

	detail := &ComponentDetail{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "ID":
			detail.ID = parts[1]
		case "Kind":
			detail.Kind = parts[1]
		case "Status":
			detail.Status = parts[1]
		case "ErrorReason":
			detail.ErrorReason = parts[1]
		}
	}
	return detail, nil
}

package drasi

import (
	"context"
	"fmt"
	"strings"

	"github.com/azure/azd.extensions.drasi/internal/output"
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
				return nil, fmt.Errorf("%s: drasi %s: %s", output.ERR_DRASI_CLI_ERROR, strings.Join(argsWithoutContext, " "), strings.TrimSpace(stderr))
			}
		} else {
			return nil, fmt.Errorf("%s: drasi %s: %s", output.ERR_DRASI_CLI_ERROR, strings.Join(args, " "), strings.TrimSpace(stderr))
		}
	}

	if strings.TrimSpace(stdout) == "" {
		return []ComponentSummary{}, nil
	}

	trimmedStdout := strings.TrimSpace(stdout)
	if strings.HasPrefix(trimmedStdout, "Error:") {
		return nil, fmt.Errorf("%s: drasi %s: %s", output.ERR_DRASI_CLI_ERROR, strings.Join(args, " "), trimmedStdout)
	}

	lines := strings.Split(stdout, "\n")
	headerIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		headerIndex = i
		break
	}

	if headerIndex == -1 {
		return []ComponentSummary{}, nil
	}

	headerLine := strings.TrimSpace(lines[headerIndex])
	if strings.HasPrefix(strings.ToLower(headerLine), "no ") {
		return []ComponentSummary{}, nil
	}

	headerFields := strings.Fields(headerLine)
	if len(headerFields) < 3 ||
		!strings.EqualFold(headerFields[0], "ID") ||
		!strings.EqualFold(headerFields[1], "KIND") ||
		!strings.EqualFold(headerFields[2], "STATUS") {
		return nil, fmt.Errorf("unexpected drasi list output: missing or malformed header")
	}

	result := []ComponentSummary{}
	for _, line := range lines[headerIndex+1:] {
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
			Status: strings.Join(fields[2:], " "),
		})
	}
	return result, nil
}

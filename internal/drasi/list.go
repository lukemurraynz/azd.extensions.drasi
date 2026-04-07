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

	return parseListOutput(stdout, kind)
}

// parseListOutput handles both pipe-delimited and space-delimited table
// formats from the Drasi CLI. The real CLI outputs pipe-delimited tables
// with kind-specific headers:
//   - source/reaction: ID | AVAILABLE | INGRESS URL | MESSAGES
//   - query:           ID | CONTAINER | ERRORMESSAGE | HOSTNAME | STATUS
//
// Older or alternative formats may use space-delimited ID KIND STATUS columns.
func parseListOutput(stdout, kind string) ([]ComponentSummary, error) {
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

	// Detect pipe-delimited format (real Drasi CLI output).
	if strings.Contains(headerLine, "|") {
		return parsePipeDelimited(lines, headerIndex, kind)
	}

	// Fall back to space-delimited ID KIND STATUS format.
	return parseSpaceDelimited(lines, headerIndex)
}

// parsePipeDelimited parses the pipe-delimited table format from the Drasi CLI.
// It locates the ID column and the best status indicator column per kind.
func parsePipeDelimited(lines []string, headerIndex int, kind string) ([]ComponentSummary, error) {
	headerLine := lines[headerIndex]
	headers := splitPipeRow(headerLine)

	idCol := -1
	statusCol := -1
	for i, h := range headers {
		upper := strings.ToUpper(h)
		if upper == "ID" {
			idCol = i
		}
		// For queries, use the STATUS column.
		// For sources/reactions, use the AVAILABLE column.
		if upper == "STATUS" {
			statusCol = i
		}
		if upper == "AVAILABLE" && statusCol == -1 {
			statusCol = i
		}
	}

	if idCol == -1 {
		return nil, fmt.Errorf("unexpected drasi list output: missing ID column in header")
	}

	// Skip the separator line (e.g., "---+---+---") and parse data rows.
	result := []ComponentSummary{}
	for _, line := range lines[headerIndex+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip separator lines that contain only dashes and pipes.
		if isSeparatorLine(trimmed) {
			continue
		}

		cols := splitPipeRow(line)
		if idCol >= len(cols) {
			continue
		}

		id := cols[idCol]
		if id == "" {
			continue
		}

		status := ""
		if statusCol >= 0 && statusCol < len(cols) {
			status = cols[statusCol]
		}

		result = append(result, ComponentSummary{
			ID:     id,
			Kind:   kind,
			Status: status,
		})
	}

	return result, nil
}

// parseSpaceDelimited handles the legacy space-delimited ID KIND STATUS format.
func parseSpaceDelimited(lines []string, headerIndex int) ([]ComponentSummary, error) {
	headerLine := strings.TrimSpace(lines[headerIndex])
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

// splitPipeRow splits a pipe-delimited row into trimmed column values.
func splitPipeRow(line string) []string {
	parts := strings.Split(line, "|")
	cols := make([]string, len(parts))
	for i, p := range parts {
		cols[i] = strings.TrimSpace(p)
	}
	return cols
}

// isSeparatorLine returns true for lines like "---+---+---" or "------+------".
func isSeparatorLine(line string) bool {
	for _, ch := range line {
		if ch != '-' && ch != '+' && ch != ' ' {
			return false
		}
	}
	return true
}

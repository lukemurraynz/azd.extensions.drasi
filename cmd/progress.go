package cmd

import (
	"time"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

// ProgressHelper wraps a yacspin spinner for CLI progress feedback.
// When the output format is JSON it becomes a no-op so that only
// machine-readable output appears on stdout.  All spinner output
// goes to stderr (via cmd.ErrOrStderr()) to keep stdout clean for
// structured command results.
type ProgressHelper struct {
	spinner *yacspin.Spinner
	noop    bool
}

// NewProgressHelper creates a ProgressHelper for the given command.
// If the command's --output flag is "json", no spinner is created and
// all methods become silent no-ops.
func NewProgressHelper(cmd *cobra.Command) (*ProgressHelper, error) {
	format := outputFormatFromCommand(cmd)
	if format == output.FormatJSON {
		return &ProgressHelper{noop: true}, nil
	}

	cfg := yacspin.Config{
		Frequency:  100 * time.Millisecond,
		CharSet:    yacspin.CharSets[14],
		Writer:     cmd.ErrOrStderr(),
		ShowCursor: true,
	}

	s, err := yacspin.New(cfg)
	if err != nil {
		return nil, err
	}

	return &ProgressHelper{spinner: s}, nil
}

// Start begins the spinner animation. No-op in JSON mode.
func (p *ProgressHelper) Start() error {
	if p.noop {
		return nil
	}
	return p.spinner.Start()
}

// Message updates the text displayed next to the spinner. No-op in JSON mode.
func (p *ProgressHelper) Message(msg string) {
	if p.noop {
		return
	}
	p.spinner.Message(" " + msg)
}

// Stop halts the spinner. No-op in JSON mode.
func (p *ProgressHelper) Stop() error {
	if p.noop {
		return nil
	}
	return p.spinner.Stop()
}

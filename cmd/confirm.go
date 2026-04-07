package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/mattn/go-isatty"
	"os"
)

// confirmFunc is a package-level var so tests can override the survey prompt.
var confirmFunc = func(prompt string, result *bool) error {
	return survey.AskOne(&survey.Confirm{Message: prompt}, result)
}

// isTTYFunc is a package-level var so tests can override TTY detection.
var isTTYFunc = func() bool {
	fd := os.Stdin.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// ConfirmDestructive checks if the user wants to proceed with a destructive action.
// If force is true: returns (true, nil) immediately without prompting.
// If stdin is not a TTY: returns (false, error) with "--force" hint.
// Otherwise: prompts user with survey.Confirm using confirmFunc.
func ConfirmDestructive(prompt string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	if !isTTYFunc() {
		return false, &commandError{
			message:  "use --force for non-interactive execution",
			exitCode: output.ExitCodes[output.ERR_FORCE_REQUIRED],
		}
	}

	var result bool
	if err := confirmFunc(prompt, &result); err != nil {
		return false, fmt.Errorf("prompt: %w", err)
	}
	return result, nil
}

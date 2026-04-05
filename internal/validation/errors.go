package validation

// ValidationLevel indicates severity of a validation issue.
type ValidationLevel string

const (
	LevelError   ValidationLevel = "error"
	LevelWarning ValidationLevel = "warning"
)

// ValidationIssue represents a single finding from the validator pipeline.
type ValidationIssue struct {
	Level       ValidationLevel `json:"level"`
	File        string          `json:"file"`
	Line        int             `json:"line"`
	Code        string          `json:"code"`
	Message     string          `json:"message"`
	Remediation string          `json:"remediation,omitempty"`
}

// ValidationResult accumulates issues from all validator stages.
type ValidationResult struct {
	Issues []ValidationIssue
}

// HasErrors returns true if any issue has LevelError.
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Level == LevelError {
			return true
		}
	}

	return false
}

// HasWarnings returns true if any issue has LevelWarning.
func (r *ValidationResult) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.Level == LevelWarning {
			return true
		}
	}

	return false
}

// Add appends an issue to the result.
func (r *ValidationResult) Add(issue ValidationIssue) {
	r.Issues = append(r.Issues, issue)
}

package config_test

import (
	"encoding/json"
	"fmt"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"testing"
)

func TestDebugSchema2(t *testing.T) {
	src := config.Source{
		APIVersion: "v1",
		Kind:       "Source",
		ID:         "my-source",
		SourceKind: "PostgreSQL",
		FilePath:   "sources/my-source.yaml",
		Line:       1,
	}
	data, _ := json.MarshalIndent(src, "", "  ")
	t.Logf("JSON: %s", string(data))

	result := &validation.ValidationResult{}
	validation.ValidateSourceSchema(src, result)
	for _, issue := range result.Issues {
		fmt.Printf("Issue: level=%s code=%s msg=%s\n", issue.Level, issue.Code, issue.Message)
		t.Logf("Issue: level=%s code=%s msg=%s", issue.Level, issue.Code, issue.Message)
	}
}

package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scaffold copies the named template into targetDir, returning the relative
// forward-slash paths of all created files.
//
// azure.yaml is silently skipped when it already exists because `azd init`
// creates it before the extension runs and binds it to the user's chosen
// project name and environment. Overwriting it would break that binding.
//
// For all other files, if force is false and the output file already exists,
// Scaffold returns an error whose message contains "already exists" without
// writing any files. If force is true, existing files are overwritten.
func Scaffold(templateName, targetDir string, force bool) ([]string, error) {
	prefix := "templates/" + templateName

	// Verify the template exists in the embedded FS before touching the filesystem.
	if _, err := fs.Stat(templateFS, prefix); err != nil {
		return nil, fmt.Errorf("template %q not found", templateName)
	}

	var results []string

	walkErr := fs.WalkDir(templateFS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Strip the "templates/<name>/" prefix to get the relative destination path.
		rel := strings.TrimPrefix(path, prefix+"/")
		outputPath := filepath.Join(targetDir, rel)

		// Path traversal guard: ensure the resolved path stays within targetDir.
		absTarget, absErr := filepath.Abs(targetDir)
		if absErr != nil {
			return fmt.Errorf("resolving target directory: %w", absErr)
		}
		absPath, absPathErr := filepath.Abs(outputPath)
		if absPathErr != nil {
			return fmt.Errorf("resolving output path %s: %w", outputPath, absPathErr)
		}
		relCheck, relErr := filepath.Rel(absTarget, absPath)
		if relErr != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("path %s escapes target directory", outputPath)
		}

		// Conflict check — azure.yaml is skipped when it already exists
		// because `azd init` creates it with the user's project name and
		// environment binding; overwriting it would break that association.
		if _, statErr := os.Stat(outputPath); statErr == nil && !force {
			if filepath.Base(outputPath) == "azure.yaml" {
				// Skip silently — the user's azure.yaml is authoritative.
				return nil
			}
			return fmt.Errorf("%s already exists", outputPath)
		}

		// Ensure parent directories exist.
		if mkErr := os.MkdirAll(filepath.Dir(outputPath), 0750); mkErr != nil {
			return fmt.Errorf("creating directory for %s: %w", outputPath, mkErr)
		}

		data, readErr := fs.ReadFile(templateFS, path)
		if readErr != nil {
			return fmt.Errorf("reading template file %s: %w", path, readErr)
		}
		if writeErr := os.WriteFile(outputPath, data, 0600); writeErr != nil {
			return fmt.Errorf("writing %s: %w", outputPath, writeErr)
		}

		// Return the relative path using forward slashes for cross-platform consistency.
		results = append(results, filepath.ToSlash(rel))
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return results, nil
}

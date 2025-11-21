package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// RegisterInModulesFile adds an import statement to the modules.go file
// to automatically register the generated toolset.
func RegisterInModulesFile(modulesFilePath, importPath string) error {
	// Read existing modules.go file
	content, err := os.ReadFile(modulesFilePath)
	if err != nil {
		return fmt.Errorf("failed to read modules.go: %w", err)
	}

	// Check if import already exists
	importLine := fmt.Sprintf("import _ %q", importPath)
	if strings.Contains(string(content), importLine) {
		// Already registered, nothing to do
		return nil
	}

	// Find the last import statement
	lines := strings.Split(string(content), "\n")
	lastImportIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import _") {
			lastImportIdx = i
		}
	}

	if lastImportIdx == -1 {
		return fmt.Errorf("no import statements found in modules.go")
	}

	// Insert new import after the last import
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:lastImportIdx+1]...)
	result = append(result, importLine)
	result = append(result, lines[lastImportIdx+1:]...)
	newContent := strings.Join(result, "\n")

	// Format the Go code
	formatted, err := format.Source([]byte(newContent))
	if err != nil {
		// If formatting fails, use unformatted content
		formatted = []byte(newContent)
	}

	// Write back to file
	if err := os.WriteFile(modulesFilePath, formatted, 0o644); err != nil {
		return fmt.Errorf("failed to write modules.go: %w", err)
	}

	return nil
}

// DetermineModulesFilePath determines the path to modules.go based on the output directory
// and module path. If modulesPath is provided, it uses that. Otherwise, it tries to
// infer the location from the output directory.
func DetermineModulesFilePath(outputDir, modulePath, modulesPath string) (string, error) {
	if modulesPath != "" {
		// Use provided path
		if !filepath.IsAbs(modulesPath) {
			return "", fmt.Errorf("modules-file path must be absolute: %s", modulesPath)
		}
		return modulesPath, nil
	}

	// Try to infer from output directory
	// Expected pattern: <repo-root>/pkg/<package-name>
	// modules.go should be at: <repo-root>/pkg/mcp/modules.go

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for output: %w", err)
	}

	// Look for pkg/ directory in the path
	parts := strings.Split(absOutputDir, string(filepath.Separator))
	pkgIdx := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "pkg" {
			pkgIdx = i
			break
		}
	}

	if pkgIdx == -1 {
		return "", fmt.Errorf("cannot infer modules.go location: output directory does not contain 'pkg' directory")
	}

	// Construct path to modules.go
	repoRoot := filepath.Join(parts[:pkgIdx]...)
	if repoRoot == "" {
		repoRoot = "/"
	}
	modulesPath = filepath.Join(string(filepath.Separator)+repoRoot, "pkg", "mcp", "modules.go")

	// Check if file exists
	if _, err := os.Stat(modulesPath); err != nil {
		return "", fmt.Errorf("modules.go not found at inferred path %s: %w", modulesPath, err)
	}

	return modulesPath, nil
}

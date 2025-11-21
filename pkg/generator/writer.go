package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// FileWriter handles writing generated code to files
type FileWriter struct {
	outputDir      string
	overwriteFiles bool
	formatCode     bool
}

// NewFileWriter creates a new FileWriter
func NewFileWriter(outputDir string, overwriteFiles, formatCode bool) *FileWriter {
	return &FileWriter{
		outputDir:      outputDir,
		overwriteFiles: overwriteFiles,
		formatCode:     formatCode,
	}
}

// WriteFile writes content to a file with optional Go formatting
func (w *FileWriter) WriteFile(filename, content string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(w.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filePath := filepath.Join(w.outputDir, filename)

	// Check if file exists and we're not overwriting
	if !w.overwriteFiles {
		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("file %s already exists and overwrite is disabled", filePath)
		}
	}

	// Format Go code if requested and filename ends with .go
	finalContent := content
	if w.formatCode && strings.HasSuffix(filename, ".go") {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			// If formatting fails, write the original content and log a warning
			fmt.Printf("Warning: failed to format %s: %v\n", filename, err)
		} else {
			finalContent = string(formatted)
		}
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(finalContent), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// WriteFiles writes multiple files
func (w *FileWriter) WriteFiles(files map[string]string) error {
	for filename, content := range files {
		if err := w.WriteFile(filename, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}
	return nil
}

// EnsureDirectory ensures that a directory exists
func (w *FileWriter) EnsureDirectory(dir string) error {
	fullPath := filepath.Join(w.outputDir, dir)
	return os.MkdirAll(fullPath, 0o755)
}

// FileExists checks if a file exists in the output directory
func (w *FileWriter) FileExists(filename string) bool {
	filePath := filepath.Join(w.outputDir, filename)
	_, err := os.Stat(filePath)
	return err == nil
}

// ReadFile reads a file from the output directory
func (w *FileWriter) ReadFile(filename string) (string, error) {
	filePath := filepath.Join(w.outputDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// ListFiles lists all files in the output directory
func (w *FileWriter) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(w.outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// Clean removes all files from the output directory
func (w *FileWriter) Clean() error {
	entries, err := os.ReadDir(w.outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(w.outputDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// GetOutputPath returns the full path for a filename in the output directory
func (w *FileWriter) GetOutputPath(filename string) string {
	return filepath.Join(w.outputDir, filename)
}

// SetOverwriteFiles sets whether to overwrite existing files
func (w *FileWriter) SetOverwriteFiles(overwrite bool) {
	w.overwriteFiles = overwrite
}

// SetFormatCode sets whether to format Go code
func (w *FileWriter) SetFormatCode(formatCode bool) {
	w.formatCode = formatCode
}

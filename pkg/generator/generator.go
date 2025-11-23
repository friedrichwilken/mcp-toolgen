// Package generator provides code generation functionality for creating Go toolsets from CRD information.
// It uses template-based generation to create complete MCP-compatible toolsets with CRUD operations.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/friedrichwilken/mcp-toolgen/pkg/analyzer"
)

// Generator handles the generation of Go code from CRD analysis
type Generator struct {
	config    *GeneratorConfig
	templates *template.Template
}

// GeneratorConfig holds configuration for code generation
type GeneratorConfig struct {
	OutputDir       string
	TemplateDir     string
	PackageName     string
	ModulePath      string
	OverwriteFiles  bool
	IncludeComments bool
}

// NewGenerator creates a new code generator
func NewGenerator(config *GeneratorConfig) (*Generator, error) {
	if config == nil {
		return nil, fmt.Errorf("generator config is required")
	}

	if config.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	generator := &Generator{
		config: config,
	}

	// Load templates
	if err := generator.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return generator, nil
}

// GenerateToolset generates a complete toolset from CRD information
func (g *Generator) GenerateToolset(toolsetInfo *analyzer.ToolsetInfo) error {
	if toolsetInfo == nil {
		return fmt.Errorf("toolset info is required")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate each file
	files := []struct {
		template string
		filename string
	}{
		{"toolset.go.tmpl", "toolset.go"},
		{"types.go.tmpl", "types.go"},
		{"client.go.tmpl", "client.go"},
		{"handlers.go.tmpl", "handlers.go"},
		{"schema.go.tmpl", "schema.go"},
		{"doc.go.tmpl", "doc.go"},
	}

	for _, file := range files {
		if err := g.generateFile(toolsetInfo, file.template, file.filename); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.filename, err)
		}
	}

	return nil
}

// generateFile generates a single file from a template
func (g *Generator) generateFile(toolsetInfo *analyzer.ToolsetInfo, templateName, filename string) error {
	// Check if file exists and we're not overwriting
	outputPath := filepath.Join(g.config.OutputDir, filename)
	if !g.config.OverwriteFiles {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("file %s already exists and overwrite is disabled", outputPath)
		}
	}

	// Execute template
	tmpl := g.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Create template data
	data := g.createTemplateData(toolsetInfo)

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return nil
}

// createTemplateData creates the data structure passed to templates
func (g *Generator) createTemplateData(toolsetInfo *analyzer.ToolsetInfo) map[string]interface{} {
	return map[string]interface{}{
		"Package":             g.config.PackageName,
		"ModulePath":          g.config.ModulePath,
		"IncludeComments":     g.config.IncludeComments,
		"GenerateCRDResource": toolsetInfo.Config.GenerateCRDResource,
		"Toolset":             toolsetInfo,
		"CRD":                 toolsetInfo.CRD,
		"MainType":            toolsetInfo.MainType,
		"SpecType":            toolsetInfo.SpecType,
		"StatusType":          toolsetInfo.StatusType,
		"ListType":            toolsetInfo.ListType,
		"Operations":          toolsetInfo.GetResourceOperations(),
		"Imports":             toolsetInfo.GetImports(),
		"KubernetesImports":   toolsetInfo.GetKubernetesImports(),
		"MCPImports":          toolsetInfo.GetMCPImports(),

		// Helper functions for templates
		"ToLower":               toLower,
		"ToUpper":               toUpper,
		"ToTitle":               toTitle,
		"ToCamelCase":           toCamelCase,
		"ToSnakeCase":           toSnakeCase,
		"Pluralize":             pluralize,
		"Contains":              contains,
		"Join":                  join,
		"Quote":                 quote,
		"ConvertSchemaToGoCode": convertSchemaToGoCode,
	}
}

// loadTemplates loads all template files
func (g *Generator) loadTemplates() error {
	templateDir := g.config.TemplateDir
	if templateDir == "" {
		// Use embedded templates (we'll implement this)
		return g.loadEmbeddedTemplates()
	}

	// Load templates from directory
	pattern := filepath.Join(templateDir, "*.tmpl")
	templates, err := template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates from %s: %w", pattern, err)
	}

	g.templates = templates
	return nil
}

// ValidateConfig validates the generator configuration
func (g *Generator) ValidateConfig() error {
	if g.config.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if g.config.PackageName == "" {
		return fmt.Errorf("package name is required")
	}

	if g.config.ModulePath == "" {
		return fmt.Errorf("module path is required")
	}

	return nil
}

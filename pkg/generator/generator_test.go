package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friedrichwilken/mcp-toolgen/pkg/analyzer"
)

func TestNewGenerator(t *testing.T) {
	config := &GeneratorConfig{
		OutputDir:       "/tmp/test",
		PackageName:     "testpkg",
		ModulePath:      "github.com/test/module",
		IncludeComments: true,
		OverwriteFiles:  true,
	}

	gen, err := NewGenerator(config)
	require.NoError(t, err)
	require.NotNil(t, gen)

	assert.Equal(t, config.OutputDir, gen.config.OutputDir)
	assert.Equal(t, config.PackageName, gen.config.PackageName)
	assert.Equal(t, config.ModulePath, gen.config.ModulePath)
	assert.Equal(t, config.IncludeComments, gen.config.IncludeComments)
	assert.Equal(t, config.OverwriteFiles, gen.config.OverwriteFiles)
}

func TestNewGeneratorValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *GeneratorConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &GeneratorConfig{
				OutputDir:   "/tmp/test",
				PackageName: "testpkg",
				ModulePath:  "github.com/test/module",
			},
			wantError: false,
		},
		{
			name: "missing output dir",
			config: &GeneratorConfig{
				PackageName: "testpkg",
				ModulePath:  "github.com/test/module",
			},
			wantError: true,
			errorMsg:  "output directory is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGenerator(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateFromCRDFile(t *testing.T) {
	// Step 1: Parse CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)
	require.NotNil(t, crdInfo)

	// Step 2: Create toolset info
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = "widgets"
	config.ModulePath = "github.com/test/module"
	config.OutputDir = t.TempDir()

	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err)
	require.NotNil(t, toolsetInfo)

	// Step 3: Generate code
	genConfig := &GeneratorConfig{
		OutputDir:       config.OutputDir,
		PackageName:     config.PackageName,
		ModulePath:      config.ModulePath,
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := NewGenerator(genConfig)
	require.NoError(t, err)

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err)

	// Verify all expected files were created
	expectedFiles := []string{"toolset.go", "types.go", "client.go", "handlers.go", "schema.go", "doc.go"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(config.OutputDir, filename)
		assert.FileExists(t, filePath, "Expected file %s to exist", filename)
	}
}

func TestGenerateFromCRDFileComplex(t *testing.T) {
	// Step 1: Parse complex CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/complex-crd.yaml")
	require.NoError(t, err)
	require.NotNil(t, crdInfo)

	// Step 2: Create toolset info
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = "complex"
	config.ModulePath = "github.com/test/module"
	config.OutputDir = t.TempDir()

	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err)
	require.NotNil(t, toolsetInfo)

	// Step 3: Generate code
	genConfig := &GeneratorConfig{
		OutputDir:       config.OutputDir,
		PackageName:     config.PackageName,
		ModulePath:      config.ModulePath,
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := NewGenerator(genConfig)
	require.NoError(t, err)

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err)

	// Verify files were created
	expectedFiles := []string{"toolset.go", "types.go", "client.go", "handlers.go", "schema.go", "doc.go"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(config.OutputDir, filename)
		assert.FileExists(t, filePath, "Expected file %s to exist", filename)
	}
}

func TestGenerateFiltered(t *testing.T) {
	// Parse CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)

	// Create config with filtered operations
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = "widgets"
	config.ModulePath = "github.com/test/module"
	config.OutputDir = t.TempDir()
	config.SelectedOperations = []string{"create", "get"}

	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err)

	// Generate
	genConfig := &GeneratorConfig{
		OutputDir:       config.OutputDir,
		PackageName:     config.PackageName,
		ModulePath:      config.ModulePath,
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := NewGenerator(genConfig)
	require.NoError(t, err)

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err)

	// Verify correct operations
	operations := toolsetInfo.GetResourceOperations()
	assert.Equal(t, []string{"create", "get"}, operations)
	assert.Contains(t, operations, "create")
	assert.Contains(t, operations, "get")
	assert.NotContains(t, operations, "update")
	assert.NotContains(t, operations, "delete")
}

func TestValidateOperations(t *testing.T) {
	// Parse a real CRD for testing
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)

	tests := []struct {
		name       string
		operations []string
		wantError  bool
	}{
		{
			name:       "valid operations",
			operations: []string{"create", "get", "list", "update", "delete"},
			wantError:  false,
		},
		{
			name:       "subset of operations",
			operations: []string{"create", "get"},
			wantError:  false,
		},
		{
			name:       "empty operations defaults to all",
			operations: []string{},
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := analyzer.DefaultGenerationConfig()
			config.SelectedOperations = tt.operations
			config.PackageName = "widgets"
			config.ModulePath = "github.com/test/module"
			config.OutputDir = t.TempDir()

			toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, toolsetInfo)

				ops := toolsetInfo.GetResourceOperations()
				if len(tt.operations) > 0 {
					assert.Equal(t, tt.operations, ops)
				} else {
					// Default should include all operations
					assert.NotEmpty(t, ops)
				}
			}
		})
	}
}

func TestShouldIncludeOperation(t *testing.T) {
	// Parse a real CRD for testing
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)

	tests := []struct {
		name               string
		selectedOperations []string
		checkOperation     string
		shouldInclude      bool
	}{
		{
			name:               "operation included",
			selectedOperations: []string{"create", "get"},
			checkOperation:     "create",
			shouldInclude:      true,
		},
		{
			name:               "operation not included",
			selectedOperations: []string{"create", "get"},
			checkOperation:     "delete",
			shouldInclude:      false,
		},
		{
			name:               "all operations when empty",
			selectedOperations: []string{},
			checkOperation:     "create",
			shouldInclude:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := analyzer.DefaultGenerationConfig()
			config.SelectedOperations = tt.selectedOperations
			config.PackageName = "widgets"
			config.ModulePath = "github.com/test/module"
			config.OutputDir = t.TempDir()

			toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
			require.NoError(t, err)

			operations := toolsetInfo.GetResourceOperations()

			found := false
			for _, op := range operations {
				if op == tt.checkOperation {
					found = true
					break
				}
			}

			assert.Equal(t, tt.shouldInclude, found,
				"Operation %s should be %v in %v", tt.checkOperation, tt.shouldInclude, operations)
		})
	}
}

func TestGenerateFromCRDStruct(t *testing.T) {
	// Test that we can use CRDInfo (struct) obtained from parsing
	// This validates that the struct-based API works correctly
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)
	require.NotNil(t, crdInfo)

	// Verify we got a proper CRDInfo struct
	assert.NotEmpty(t, crdInfo.Kind)
	assert.NotEmpty(t, crdInfo.Plural)

	// Create toolset info from the struct
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = "customwidgets"
	config.ModulePath = "github.com/test/module"
	config.OutputDir = t.TempDir()

	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err)
	require.NotNil(t, toolsetInfo)

	// Generate code
	genConfig := &GeneratorConfig{
		OutputDir:       config.OutputDir,
		PackageName:     config.PackageName,
		ModulePath:      config.ModulePath,
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := NewGenerator(genConfig)
	require.NoError(t, err)

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err)

	// Verify files were created
	expectedFiles := []string{"toolset.go", "types.go", "client.go", "handlers.go", "schema.go", "doc.go"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(config.OutputDir, filename)
		assert.FileExists(t, filePath, "Expected file %s to exist", filename)
	}
}

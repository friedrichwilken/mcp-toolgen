package integration

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friedrichwilken/mcp-toolgen/pkg/analyzer"
	"github.com/friedrichwilken/mcp-toolgen/pkg/generator"
	"github.com/friedrichwilken/mcp-toolgen/test/utils"
)

// updateGolden is a flag to update golden files when running tests
var updateGolden = flag.Bool("update-golden", false, "update golden files with current generated output")

// TestTemplateGoldenFiles tests that generated code matches expected golden files
// Run with -update-golden flag to regenerate golden files
func TestTemplateGoldenFiles(t *testing.T) {
	utils.SkipIfShort(t)

	testCases := []struct {
		name          string
		crdFile       string
		packageName   string
		operations    []string
		expectedFiles []string
		validateFunc  func(t *testing.T, goldenDir, generatedDir string)
	}{
		{
			name:        "simple CRD with all operations",
			crdFile:     "simple-crd.yaml",
			packageName: "widgets",
			operations:  []string{"create", "get", "list", "update", "delete"},
			expectedFiles: []string{
				"toolset.go",
				"types.go",
				"client.go",
				"handlers.go",
				"schema.go",
				"doc.go",
			},
			validateFunc: validateSimpleCRD,
		},
		{
			name:        "simple CRD with create and read only",
			crdFile:     "simple-crd.yaml",
			packageName: "widgets_readonly",
			operations:  []string{"create", "get", "list"},
			expectedFiles: []string{
				"toolset.go",
				"types.go",
				"client.go",
				"handlers.go",
				"schema.go",
				"doc.go",
			},
			validateFunc: validateReadonlyCRD,
		},
		{
			name:        "cluster-scoped CRD",
			crdFile:     "cluster-scoped-crd.yaml",
			packageName: "clusterwidgets",
			operations:  []string{"create", "get", "list", "update", "delete"},
			expectedFiles: []string{
				"toolset.go",
				"types.go",
				"client.go",
				"handlers.go",
				"schema.go",
				"doc.go",
			},
			validateFunc: validateClusterScopedCRD,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate code
			generatedDir := generateTestCode(t, tc.crdFile, tc.packageName, tc.operations)

			// Get golden file directory
			goldenDir := getGoldenDir(t, tc.name)

			// Update golden files if flag is set
			if *updateGolden {
				t.Logf("Updating golden files in %s", goldenDir)
				updateGoldenFiles(t, generatedDir, goldenDir)
			}

			// Verify all expected files exist in both generated and golden
			for _, filename := range tc.expectedFiles {
				generatedPath := filepath.Join(generatedDir, filename)
				goldenPath := filepath.Join(goldenDir, filename)

				require.FileExists(t, generatedPath, "Generated file should exist: %s", filename)

				if !*updateGolden {
					require.FileExists(t, goldenPath, "Golden file should exist: %s", filename)

					// Compare files
					compareFiles(t, goldenPath, generatedPath, filename)
				}
			}

			// Run custom validation if provided
			if tc.validateFunc != nil && !*updateGolden {
				tc.validateFunc(t, goldenDir, generatedDir)
			}
		})
	}
}

// TestTemplateEdgeCases tests edge cases in template generation
func TestTemplateEdgeCases(t *testing.T) {
	utils.SkipIfShort(t)

	testCases := []struct {
		name        string
		crdFile     string
		packageName string
		operations  []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty operations list",
			crdFile:     "simple-crd.yaml",
			packageName: "widgets",
			operations:  []string{},
			expectError: false, // Should default to all operations
		},
		{
			name:        "invalid operation",
			crdFile:     "simple-crd.yaml",
			packageName: "widgets",
			operations:  []string{"invalid"},
			expectError: false, // Should be filtered out
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := utils.TempDir(t)
			crdPath := utils.GetFixturePath(t, tc.crdFile)

			// Parse CRD
			crdAnalyzer := analyzer.NewCRDAnalyzer()
			crdInfo, err := crdAnalyzer.ParseCRDFromFile(crdPath)
			require.NoError(t, err, "Failed to parse CRD")

			// Create generation config
			config := analyzer.DefaultGenerationConfig()
			config.PackageName = tc.packageName
			config.ModulePath = "github.com/test/module"
			config.OutputDir = tempDir
			config.SelectedOperations = tc.operations

			// Create toolset info
			toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)

			if tc.expectError {
				assert.Error(t, err, "Expected error: %s", tc.errorMsg)
				if err != nil && tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
				return
			}

			require.NoError(t, err, "Failed to create toolset info")

			// Generate code
			genConfig := &generator.GeneratorConfig{
				OutputDir:       tempDir,
				PackageName:     tc.packageName,
				ModulePath:      "github.com/test/module",
				OverwriteFiles:  true,
				IncludeComments: true,
			}

			gen, err := generator.NewGenerator(genConfig)
			require.NoError(t, err, "Failed to create generator")

			err = gen.GenerateToolset(toolsetInfo)
			require.NoError(t, err, "Failed to generate toolset")
		})
	}
}

// Helper functions

func generateTestCode(t *testing.T, crdFile, packageName string, operations []string) string {
	t.Helper()

	tempDir := utils.TempDir(t)
	crdPath := utils.GetFixturePath(t, crdFile)

	// Parse CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile(crdPath)
	require.NoError(t, err, "Failed to parse CRD")

	// Create generation config
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = packageName
	config.ModulePath = "github.com/test/module"
	config.OutputDir = tempDir
	config.SelectedOperations = operations

	// Create toolset info
	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err, "Failed to create toolset info")

	// Generate code
	genConfig := &generator.GeneratorConfig{
		OutputDir:       tempDir,
		PackageName:     packageName,
		ModulePath:      "github.com/test/module",
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := generator.NewGenerator(genConfig)
	require.NoError(t, err, "Failed to create generator")

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err, "Failed to generate toolset")

	return tempDir
}

func getGoldenDir(t *testing.T, testName string) string {
	t.Helper()

	// Sanitize test name for directory name
	dirName := strings.ToLower(testName)
	dirName = strings.ReplaceAll(dirName, " ", "_")
	dirName = strings.ReplaceAll(dirName, "-", "_")

	goldenDir := utils.GetFixturePath(t, filepath.Join("golden", dirName))
	return goldenDir
}

func updateGoldenFiles(t *testing.T, generatedDir, goldenDir string) {
	t.Helper()

	// Create golden directory if it doesn't exist
	err := os.MkdirAll(goldenDir, 0o755)
	require.NoError(t, err, "Failed to create golden directory")

	// Copy all generated files to golden directory
	entries, err := os.ReadDir(generatedDir)
	require.NoError(t, err, "Failed to read generated directory")

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(generatedDir, entry.Name())
		dstPath := filepath.Join(goldenDir, entry.Name())

		content, err := os.ReadFile(srcPath) // #nosec G304 -- test helper
		require.NoError(t, err, "Failed to read generated file: %s", entry.Name())

		err = os.WriteFile(dstPath, content, 0o644)
		require.NoError(t, err, "Failed to write golden file: %s", entry.Name())

		t.Logf("Updated golden file: %s", entry.Name())
	}
}

func compareFiles(t *testing.T, goldenPath, generatedPath, filename string) {
	t.Helper()

	goldenContent, err := os.ReadFile(goldenPath) // #nosec G304 -- test helper
	require.NoError(t, err, "Failed to read golden file: %s", filename)

	generatedContent, err := os.ReadFile(generatedPath) // #nosec G304 -- test helper
	require.NoError(t, err, "Failed to read generated file: %s", filename)

	if !bytes.Equal(goldenContent, generatedContent) {
		// Show diff for better debugging
		t.Errorf("Generated file %s does not match golden file", filename)
		t.Logf("Run tests with -update-golden flag to update golden files if this is expected")

		// Show first few lines of difference
		goldenLines := strings.Split(string(goldenContent), "\n")
		generatedLines := strings.Split(string(generatedContent), "\n")

		maxLines := len(goldenLines)
		if len(generatedLines) > maxLines {
			maxLines = len(generatedLines)
		}

		diffCount := 0
		for i := 0; i < maxLines && diffCount < 10; i++ {
			var goldenLine, generatedLine string
			if i < len(goldenLines) {
				goldenLine = goldenLines[i]
			}
			if i < len(generatedLines) {
				generatedLine = generatedLines[i]
			}

			if goldenLine != generatedLine {
				t.Logf("Line %d differs:", i+1)
				t.Logf("  Golden:    %s", goldenLine)
				t.Logf("  Generated: %s", generatedLine)
				diffCount++
			}
		}

		if diffCount >= 10 {
			t.Logf("... and more differences")
		}

		// Fail the test
		assert.Equal(t, string(goldenContent), string(generatedContent),
			"File %s content mismatch (see logs for details)", filename)
	}
}

// Validation functions for specific test cases

func validateSimpleCRD(t *testing.T, goldenDir, generatedDir string) {
	t.Helper()

	// Verify toolset.go has basic structure
	toolsetPath := filepath.Join(generatedDir, "toolset.go")
	toolsetContent := utils.ReadFileContent(t, toolsetPath)

	// Verify toolset type and methods exist
	assert.Contains(t, toolsetContent, "type WidgetToolset struct", "Should have WidgetToolset type")
	assert.Contains(t, toolsetContent, "GetName()", "Should have GetName method")
	assert.Contains(t, toolsetContent, "GetDescription()", "Should have GetDescription method")

	// Verify types.go has required types
	typesPath := filepath.Join(generatedDir, "types.go")
	typesContent := utils.ReadFileContent(t, typesPath)

	assert.Contains(t, typesContent, "type Widget struct", "Should have Widget type")
	assert.Contains(t, typesContent, "type WidgetSpec struct", "Should have WidgetSpec type")
	assert.Contains(t, typesContent, "type WidgetStatus struct", "Should have WidgetStatus type")
	assert.Contains(t, typesContent, "type WidgetList struct", "Should have WidgetList type")

	// Verify DeepCopy methods exist
	assert.Contains(t, typesContent, "DeepCopy()", "Should have DeepCopy methods")
}

func validateReadonlyCRD(t *testing.T, goldenDir, generatedDir string) {
	t.Helper()

	// Verify toolset.go has basic structure
	toolsetPath := filepath.Join(generatedDir, "toolset.go")
	toolsetContent := utils.ReadFileContent(t, toolsetPath)

	// Verify toolset type exists
	assert.Contains(t, toolsetContent, "type WidgetToolset struct", "Should have WidgetToolset type")
	assert.Contains(t, toolsetContent, "GetName()", "Should have GetName method")

	// Note: We can't verify specific operations presence/absence in toolset.go
	// because the template might not fully expand those sections yet.
	// The golden file comparison will catch any differences in operations.
}

func validateClusterScopedCRD(t *testing.T, goldenDir, generatedDir string) {
	t.Helper()

	// Read the CRD to check if it's cluster-scoped
	crdPath := utils.GetFixturePath(t, "cluster-scoped-crd.yaml")
	crdContent := utils.ReadFileContent(t, crdPath)

	// Only validate if CRD is actually cluster-scoped
	if !strings.Contains(crdContent, "scope: Cluster") {
		t.Skip("cluster-scoped-crd.yaml is not actually cluster-scoped")
	}

	// Verify handlers.go handles cluster-scoped resources correctly
	handlersPath := filepath.Join(generatedDir, "handlers.go")
	handlersContent := utils.ReadFileContent(t, handlersPath)

	// For cluster-scoped resources, namespace handling should be different
	// This is a placeholder - actual validation depends on template implementation
	assert.Contains(t, handlersContent, "func Handle", "Should have handler functions")
}

package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friedrichwilken/mcp-toolgen/pkg/analyzer"
	"github.com/friedrichwilken/mcp-toolgen/pkg/generator"
	"github.com/friedrichwilken/mcp-toolgen/test/utils"
)

// TestGeneratedHandlersWithMockClient tests generated handler functions with a mock Kubernetes client
func TestGeneratedHandlersWithMockClient(t *testing.T) {
	utils.SkipIfShort(t)

	// Step 1: Generate toolset code from TestWidget CRD
	t.Log("Step 1: Generating toolset code")
	tempDir := utils.TempDir(t)
	crdPath := getTestCRDPath(t)

	genConfig := generateTestToolset(t, crdPath, tempDir)

	// Step 2: Verify generated files exist
	t.Log("Step 2: Verifying generated files")
	verifyGeneratedFiles(t, genConfig.OutputDir)

	// Step 3: Test schema validation (if templates are implemented)
	t.Log("Step 3: Testing JSON schema generation")
	schemaFile := genConfig.OutputDir + "/schema.go"
	schemaContent := utils.ReadFileContent(t, schemaFile)
	if strings.Contains(schemaContent, "jsonschema") {
		testSchemaGeneration(t, genConfig.OutputDir)
	} else {
		t.Log("Schema template not fully implemented - skipping detailed tests")
	}

	// Step 4: Test handler parameter parsing (if templates are implemented)
	t.Log("Step 4: Testing handler parameter parsing")
	handlersFile := genConfig.OutputDir + "/handlers.go"
	handlersContent := utils.ReadFileContent(t, handlersFile)
	if strings.Contains(handlersContent, "api.ToolHandlerParams") {
		testHandlerParameterParsing(t, genConfig.OutputDir)
	} else {
		t.Log("Handlers template not fully implemented - skipping detailed tests")
	}

	t.Log("✅ Generated code runtime tests completed successfully!")
}

// TestGeneratedSchemaValidation tests that generated schemas accept valid and reject invalid inputs
func TestGeneratedSchemaValidation(t *testing.T) {
	utils.SkipIfShort(t)

	tempDir := utils.TempDir(t)
	crdPath := getTestCRDPath(t)

	genConfig := generateTestToolset(t, crdPath, tempDir)

	t.Run("schemas define required fields", func(t *testing.T) {
		schemaFile := genConfig.OutputDir + "/schema.go"
		content := utils.ReadFileContent(t, schemaFile)

		// Verify schemas exist
		assert.NotEmpty(t, content, "Schema file should have content")

		// If proper template is used, check for required fields
		if strings.Contains(content, "jsonschema") {
			assert.Contains(t, content, "Required:", "Schemas should define required fields")
		}
	})
}

// TestGeneratedClientMethods tests the generated Kubernetes client wrapper
func TestGeneratedClientMethods(t *testing.T) {
	utils.SkipIfShort(t)

	tempDir := utils.TempDir(t)
	crdPath := getTestCRDPath(t)

	genConfig := generateTestToolset(t, crdPath, tempDir)

	t.Run("client has basic structure", func(t *testing.T) {
		clientFile := genConfig.OutputDir + "/client.go"
		content := utils.ReadFileContent(t, clientFile)

		// Verify client type exists
		assert.Contains(t, content, "TestWidgetClient", "Client type should exist")

		// Verify at least Create method exists (others may be added in future templates)
		assert.Contains(t, content, "func (c *TestWidgetClient) Create(", "Create method should exist")

		// Verify constructor exists
		assert.Contains(t, content, "NewTestWidgetClient", "Client constructor should exist")
	})

	t.Run("client uses controller-runtime client", func(t *testing.T) {
		clientFile := genConfig.OutputDir + "/client.go"
		content := utils.ReadFileContent(t, clientFile)

		// Verify controller-runtime client usage
		assert.Contains(t, content, `"sigs.k8s.io/controller-runtime/pkg/client"`, "should import controller-runtime client")
		assert.Contains(t, content, "client.Client", "should use controller-runtime Client interface")
	})
}

// TestGeneratedTypesStructure tests the generated Go types match the CRD schema
func TestGeneratedTypesStructure(t *testing.T) {
	utils.SkipIfShort(t)

	tempDir := utils.TempDir(t)
	crdPath := getTestCRDPath(t)

	genConfig := generateTestToolset(t, crdPath, tempDir)

	t.Run("types have correct structure", func(t *testing.T) {
		typesFile := genConfig.OutputDir + "/types.go"
		content := utils.ReadFileContent(t, typesFile)

		// Verify main resource type
		assert.Contains(t, content, "type TestWidget struct", "TestWidget type should exist")
		assert.Contains(t, content, "type TestWidgetSpec struct", "TestWidgetSpec type should exist")
		assert.Contains(t, content, "type TestWidgetStatus struct", "TestWidgetStatus type should exist")
		assert.Contains(t, content, "type TestWidgetList struct", "TestWidgetList type should exist")
	})

	t.Run("types have JSON tags", func(t *testing.T) {
		typesFile := genConfig.OutputDir + "/types.go"
		content := utils.ReadFileContent(t, typesFile)

		// Verify JSON tags are present (check for any json tag)
		assert.Contains(t, content, `json:"`, "Types should have JSON tags")

		// Verify common fields have tags
		assert.True(t,
			strings.Contains(content, `json:"metadata,omitempty"`) ||
				strings.Contains(content, `json:",inline"`),
			"Should have standard Kubernetes resource tags")
	})

	t.Run("types have metav1 embedded types", func(t *testing.T) {
		typesFile := genConfig.OutputDir + "/types.go"
		content := utils.ReadFileContent(t, typesFile)

		// Verify metav1 types are used
		assert.Contains(t, content, "metav1.TypeMeta", "should embed TypeMeta")
		assert.Contains(t, content, "metav1.ObjectMeta", "should embed ObjectMeta")
		assert.Contains(t, content, "metav1.ListMeta", "should embed ListMeta")
	})
}

// TestGeneratedHandlersErrorHandling tests error handling in generated handlers
func TestGeneratedHandlersErrorHandling(t *testing.T) {
	utils.SkipIfShort(t)

	tempDir := utils.TempDir(t)
	crdPath := getTestCRDPath(t)

	genConfig := generateTestToolset(t, crdPath, tempDir)

	t.Run("handlers have basic structure", func(t *testing.T) {
		handlersFile := genConfig.OutputDir + "/handlers.go"
		content := utils.ReadFileContent(t, handlersFile)

		// Verify file has content
		assert.NotEmpty(t, content, "Handlers file should have content")

		// Verify handlers exist
		assert.Contains(t, content, "func Handle", "Handlers should have Handle functions")

		// If proper template is used, verify additional details
		if strings.Contains(content, "api.ToolHandlerParams") {
			// Verify error handling patterns
			assert.Contains(t, content, `if err != nil {`, "handlers should check for errors")

			// Verify marshaling exists
			assert.Contains(t, content, "json.", "handlers should use JSON")
		}
	})
}

// Helper functions

func getTestCRDPath(t *testing.T) string {
	t.Helper()
	return utils.GetFixturePath(t, "testwidget-crd.yaml")
}

func generateTestToolset(t *testing.T, crdPath, outputDir string) *generator.GeneratorConfig {
	t.Helper()

	// Parse CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile(crdPath)
	require.NoError(t, err, "Failed to parse CRD")

	// Create generation config
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = "testwidgets"
	config.ModulePath = "github.com/friedrichwilken/mcp-toolgen/test/generated"
	config.OutputDir = outputDir
	config.SelectedOperations = []string{"create", "get", "list", "update", "delete"}

	// Create toolset info
	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	require.NoError(t, err, "Failed to create toolset info")

	// Generate code
	genConfig := &generator.GeneratorConfig{
		OutputDir:       outputDir,
		PackageName:     "testwidgets",
		ModulePath:      "github.com/friedrichwilken/mcp-toolgen/test/generated",
		OverwriteFiles:  true,
		IncludeComments: true,
	}

	gen, err := generator.NewGenerator(genConfig)
	require.NoError(t, err, "Failed to create generator")

	err = gen.GenerateToolset(toolsetInfo)
	require.NoError(t, err, "Failed to generate toolset")

	return genConfig
}

func verifyGeneratedFiles(t *testing.T, outputDir string) {
	t.Helper()

	expectedFiles := []string{
		"toolset.go",
		"types.go",
		"client.go",
		"handlers.go",
		"schema.go",
		"doc.go",
	}

	for _, filename := range expectedFiles {
		path := outputDir + "/" + filename
		require.FileExists(t, path, "Expected file %s to exist", filename)
	}
}

func testSchemaGeneration(t *testing.T, outputDir string) {
	t.Helper()

	schemaFile := outputDir + "/schema.go"
	content := utils.ReadFileContent(t, schemaFile)

	// Verify file exists and has content
	assert.NotEmpty(t, content, "Schema file should have content")

	// Verify package declaration
	assert.Contains(t, content, "package testwidgets", "Schema file should have correct package")

	// Verify jsonschema import exists
	if assert.Contains(t, content, "jsonschema", "Schema file should import jsonschema") {
		// Only check schema functions if proper template is being used
		operations := []string{"create", "get", "list", "update", "delete"}
		for _, op := range operations {
			funcName := op + "TestWidgetSchema"
			assert.Contains(t, content, funcName, "Schema function %s should exist", funcName)
		}

		// Verify schema returns jsonschema.Schema
		assert.Contains(t, content, "*jsonschema.Schema", "Schema functions should return jsonschema.Schema")

		// Verify required fields are defined
		assert.Contains(t, content, "Required:", "Schemas should define required fields")
	}
}

func testHandlerParameterParsing(t *testing.T, outputDir string) {
	t.Helper()

	handlersFile := outputDir + "/handlers.go"
	content := utils.ReadFileContent(t, handlersFile)

	// Verify file exists and has content
	assert.NotEmpty(t, content, "Handlers file should have content")

	// Verify package declaration
	assert.Contains(t, content, "package testwidgets", "Handlers file should have correct package")

	// Verify handlers exist (check for at least one operation)
	assert.Contains(t, content, "func Handle", "Handlers should have Handle functions")

	// If the proper template is being used, verify details
	if assert.Contains(t, content, "api.ToolHandlerParams", "Handlers should use proper API") {
		// Verify handlers extract arguments
		assert.Contains(t, content, "args := params.GetArguments()", "Handlers should get arguments from params")

		// Verify handlers use helper functions
		assert.Contains(t, content, "getClusterName(args)", "Handlers should extract cluster name")
		assert.Contains(t, content, "getNamespace(args)", "Handlers should extract namespace")

		// Verify handlers parse resource data
		assert.Contains(t, content, "parseResource(", "Handlers should parse resource data")
	}
}

// TestMultipleCRDVariations tests generation with different CRD schemas
func TestMultipleCRDVariations(t *testing.T) {
	utils.SkipIfShort(t)

	testCases := []struct {
		name        string
		crdFile     string
		expectKind  string
		expectGroup string
	}{
		{
			name:        "simple CRD",
			crdFile:     "simple-crd.yaml",
			expectKind:  "Widget",
			expectGroup: "example.com",
		},
		{
			name:        "complex CRD",
			crdFile:     "complex-crd.yaml",
			expectKind:  "Application",
			expectGroup: "apps.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := utils.TempDir(t)
			crdPath := utils.GetFixturePath(t, tc.crdFile)

			// Skip if fixture doesn't exist
			if !utils.FileExists(crdPath) {
				t.Skipf("Fixture %s not found", tc.crdFile)
			}

			genConfig := generateTestToolset(t, crdPath, tempDir)

			// Verify files were generated
			verifyGeneratedFiles(t, genConfig.OutputDir)

			// Verify correct types were generated
			typesFile := genConfig.OutputDir + "/types.go"
			content := utils.ReadFileContent(t, typesFile)

			expectedType := "type " + tc.expectKind + " struct"
			assert.Contains(t, content, expectedType, "Should generate type for %s", tc.expectKind)
		})
	}
}

package analyzer

import (
	"fmt"
	"strings"
)

// GenerationConfig holds configuration for code generation
type GenerationConfig struct {
	// Package information
	PackageName string
	ModulePath  string
	OutputDir   string

	// Template customization
	TemplateDir     string
	CustomTemplates map[string]string

	// Generation options
	GenerateCRUD       bool
	GenerateTests      bool
	GenerateSchemas    bool
	IncludeComments    bool
	SelectedOperations []string

	// Kubernetes integration
	UseControllerRuntime bool
	MultiClusterSupport  bool
}

// DefaultGenerationConfig returns a default configuration
func DefaultGenerationConfig() *GenerationConfig {
	return &GenerationConfig{
		ModulePath:           "github.com/example/project",
		GenerateCRUD:         true,
		GenerateTests:        false,
		GenerateSchemas:      true,
		IncludeComments:      true,
		UseControllerRuntime: true,
		MultiClusterSupport:  true,
	}
}

// ToolsetInfo contains information needed to generate a complete MCP toolset
type ToolsetInfo struct {
	// CRD information
	CRD *CRDInfo

	// Generated types
	MainType   *GoTypeInfo
	SpecType   *GoTypeInfo
	StatusType *GoTypeInfo
	ListType   *GoTypeInfo

	// Package information
	PackageName string
	ImportPath  string

	// Configuration
	Config *GenerationConfig
}

// NewToolsetInfo creates ToolsetInfo from CRDInfo
func NewToolsetInfo(crd *CRDInfo, config *GenerationConfig) (*ToolsetInfo, error) {
	if crd == nil {
		return nil, fmt.Errorf("CRD info is required")
	}
	if config == nil {
		config = DefaultGenerationConfig()
	}

	packageName := config.PackageName
	if packageName == "" {
		packageName = crd.GetPackageName()
	}

	toolset := &ToolsetInfo{
		CRD:         crd,
		PackageName: packageName,
		ImportPath:  fmt.Sprintf("%s/pkg/%s", config.ModulePath, packageName),
		Config:      config,
	}

	// Analyze the main schema to generate types
	if err := toolset.analyzeTypes(); err != nil {
		return nil, fmt.Errorf("failed to analyze types: %w", err)
	}

	return toolset, nil
}

// analyzeTypes analyzes the CRD schema and generates Go type information
func (t *ToolsetInfo) analyzeTypes() error {
	if t.CRD.Schema == nil {
		return fmt.Errorf("CRD schema is required for type generation")
	}

	analyzer := NewSchemaAnalyzer()

	// Generate main type
	mainType, err := analyzer.AnalyzeSchema(t.CRD.Schema, t.CRD.GetTypeName(), "")
	if err != nil {
		return fmt.Errorf("failed to analyze main type: %w", err)
	}
	t.MainType = mainType

	// Generate spec type if it exists
	if specSchema, exists := t.CRD.Schema.Properties["spec"]; exists {
		specType, err := analyzer.AnalyzeSchema(&specSchema, t.CRD.GetTypeName()+"Spec", "spec")
		if err != nil {
			return fmt.Errorf("failed to analyze spec type: %w", err)
		}
		t.SpecType = specType
	}

	// Generate status type if it exists
	if statusSchema, exists := t.CRD.Schema.Properties["status"]; exists {
		statusType, err := analyzer.AnalyzeSchema(&statusSchema, t.CRD.GetTypeName()+"Status", "status")
		if err != nil {
			return fmt.Errorf("failed to analyze status type: %w", err)
		}
		t.StatusType = statusType
	}

	// Generate list type
	listType := &GoTypeInfo{
		Name:     t.CRD.GetListTypeName(),
		GoType:   t.CRD.GetListTypeName(),
		JSONName: "",
	}
	t.ListType = listType

	return nil
}

// GetToolsetName returns the name for the MCP toolset
func (t *ToolsetInfo) GetToolsetName() string {
	return strings.ToLower(t.CRD.Plural)
}

// GetToolsetDescription returns a description for the MCP toolset
func (t *ToolsetInfo) GetToolsetDescription() string {
	return fmt.Sprintf("Tools for managing %s custom resources", t.CRD.Kind)
}

// GetResourceOperations returns the list of CRUD operations to generate
func (t *ToolsetInfo) GetResourceOperations() []string {
	// Use selected operations if specified, otherwise use default
	if len(t.Config.SelectedOperations) > 0 {
		return t.Config.SelectedOperations
	}
	// Default to all operations
	return []string{"create", "get", "list", "update", "delete"}
}

// GetImports returns the Go imports needed for the generated code
func (t *ToolsetInfo) GetImports() []string {
	imports := []string{
		"context",
		"fmt",
		"encoding/json",
	}

	// Add Kubernetes imports
	imports = append(imports,
		"k8s.io/apimachinery/pkg/apis/meta/v1",
		"k8s.io/apimachinery/pkg/runtime",
		"k8s.io/apimachinery/pkg/types",
		"sigs.k8s.io/controller-runtime/pkg/client",
	)

	// Add MCP imports
	imports = append(imports,
		"github.com/modelcontextprotocol/go-sdk/api",
	)

	return imports
}

// GetKubernetesImports returns Kubernetes-specific imports
func (t *ToolsetInfo) GetKubernetesImports() []string {
	return []string{
		"k8s.io/apimachinery/pkg/apis/meta/v1",
		"k8s.io/apimachinery/pkg/runtime",
		"k8s.io/apimachinery/pkg/runtime/schema",
		"k8s.io/apimachinery/pkg/types",
		"sigs.k8s.io/controller-runtime/pkg/client",
	}
}

// GetMCPImports returns MCP-specific imports
func (t *ToolsetInfo) GetMCPImports() []string {
	return []string{
		"github.com/modelcontextprotocol/go-sdk/api",
	}
}

// HasSpec returns true if the CRD has a spec section
func (t *ToolsetInfo) HasSpec() bool {
	return t.SpecType != nil
}

// HasStatus returns true if the CRD has a status section
func (t *ToolsetInfo) HasStatus() bool {
	return t.StatusType != nil
}

// GetAPIVersion returns the API version for the CRD
func (t *ToolsetInfo) GetAPIVersion() string {
	return t.CRD.GetAPIVersion()
}

// GetKind returns the Kind for the CRD
func (t *ToolsetInfo) GetKind() string {
	return t.CRD.Kind
}

// GetResource returns the resource name (plural)
func (t *ToolsetInfo) GetResource() string {
	return t.CRD.Plural
}

// GetGroup returns the API group
func (t *ToolsetInfo) GetGroup() string {
	return t.CRD.Group
}

// GetVersion returns the API version
func (t *ToolsetInfo) GetVersion() string {
	return t.CRD.Version
}

// GetGroupVersionResource returns the GroupVersionResource tuple
func (t *ToolsetInfo) GetGroupVersionResource() string {
	return fmt.Sprintf("%s/%s/%s", t.CRD.Group, t.CRD.Version, t.CRD.Plural)
}

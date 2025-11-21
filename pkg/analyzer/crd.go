// Package analyzer provides functionality for parsing and analyzing Kubernetes Custom Resource Definitions (CRDs).
// It extracts metadata, schemas, and type information from CRD YAML files to prepare them for code generation.
package analyzer

import (
	"fmt"
	"io"
	"os"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"
)

// CRDAnalyzer provides functionality to parse and analyze CustomResourceDefinitions
type CRDAnalyzer struct {
	scheme *runtime.Scheme
	codecs serializer.CodecFactory
}

// NewCRDAnalyzer creates a new CRDAnalyzer instance
func NewCRDAnalyzer() *CRDAnalyzer {
	scheme := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme) // Error is always nil for well-known schemes
	codecs := serializer.NewCodecFactory(scheme)

	return &CRDAnalyzer{
		scheme: scheme,
		codecs: codecs,
	}
}

// CRDInfo contains extracted information from a CRD
type CRDInfo struct {
	// Basic metadata
	Name     string
	Group    string
	Kind     string
	Version  string
	Versions []string

	// Resource naming
	Plural     string
	Singular   string
	ShortNames []string
	ListKind   string

	// Schema information
	Schema        *apiextensionsv1.JSONSchemaProps
	OpenAPISchema *apiextensionsv1.JSONSchemaProps

	// Original CRD for reference
	CRD *apiextensionsv1.CustomResourceDefinition
}

// ParseCRDFromFile parses a CRD from a YAML file
func (a *CRDAnalyzer) ParseCRDFromFile(filename string) (*CRDInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CRD file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	return a.ParseCRDFromReader(file)
}

// ParseCRDFromReader parses a CRD from an io.Reader
func (a *CRDAnalyzer) ParseCRDFromReader(reader io.Reader) (*CRDInfo, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRD data: %w", err)
	}

	return a.ParseCRDFromYAML(data)
}

// ParseCRDFromYAML parses a CRD from YAML bytes
func (a *CRDAnalyzer) ParseCRDFromYAML(yamlData []byte) (*CRDInfo, error) {
	// Convert YAML to JSON
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	// Decode into CRD object
	decoder := a.codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(jsonData, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CRD: %w", err)
	}

	crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
	if !ok {
		return nil, fmt.Errorf("object is not a CustomResourceDefinition, got %T", obj)
	}

	return a.AnalyzeCRD(crd)
}

// AnalyzeCRD analyzes a CRD and extracts relevant information
func (a *CRDAnalyzer) AnalyzeCRD(crd *apiextensionsv1.CustomResourceDefinition) (*CRDInfo, error) {
	if err := a.ValidateCRD(crd); err != nil {
		return nil, fmt.Errorf("CRD validation failed: %w", err)
	}

	info := &CRDInfo{
		Name:       crd.Name,
		Group:      crd.Spec.Group,
		Kind:       crd.Spec.Names.Kind,
		Plural:     crd.Spec.Names.Plural,
		Singular:   crd.Spec.Names.Singular,
		ShortNames: crd.Spec.Names.ShortNames,
		ListKind:   crd.Spec.Names.ListKind,
		CRD:        crd,
	}

	// Extract version information
	if len(crd.Spec.Versions) > 0 {
		// Get the storage version (or first version if no storage version is set)
		var storageVersion *apiextensionsv1.CustomResourceDefinitionVersion
		for i := range crd.Spec.Versions {
			version := &crd.Spec.Versions[i]
			info.Versions = append(info.Versions, version.Name)

			if version.Storage || storageVersion == nil {
				storageVersion = version
				info.Version = version.Name
			}
		}

		// Extract schema from storage version
		if storageVersion != nil && storageVersion.Schema != nil && storageVersion.Schema.OpenAPIV3Schema != nil {
			info.Schema = storageVersion.Schema.OpenAPIV3Schema
			info.OpenAPISchema = storageVersion.Schema.OpenAPIV3Schema
		}
	}

	// Set ListKind if not specified
	if info.ListKind == "" {
		info.ListKind = info.Kind + "List"
	}

	return info, nil
}

// ValidateCRD validates that a CRD has the required fields for code generation
func (a *CRDAnalyzer) ValidateCRD(crd *apiextensionsv1.CustomResourceDefinition) error {
	if crd == nil {
		return fmt.Errorf("CRD is nil")
	}

	if crd.Spec.Group == "" {
		return fmt.Errorf("CRD group is required")
	}

	if crd.Spec.Names.Kind == "" {
		return fmt.Errorf("CRD kind is required")
	}

	if crd.Spec.Names.Plural == "" {
		return fmt.Errorf("CRD plural name is required")
	}

	if len(crd.Spec.Versions) == 0 {
		return fmt.Errorf("CRD must have at least one version")
	}

	// Validate that at least one version has a schema
	hasSchema := false
	for _, version := range crd.Spec.Versions {
		if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
			hasSchema = true
			break
		}
	}

	if !hasSchema {
		return fmt.Errorf("CRD must have at least one version with an OpenAPI v3 schema")
	}

	return nil
}

// GetPackageName generates a Go package name from the CRD information
func (info *CRDInfo) GetPackageName() string {
	// Use plural name, convert to lowercase and replace hyphens with underscores
	packageName := strings.ToLower(info.Plural)
	packageName = strings.ReplaceAll(packageName, "-", "_")
	return packageName
}

// GetTypeName generates the main Go type name for the custom resource
func (info *CRDInfo) GetTypeName() string {
	return info.Kind
}

// GetListTypeName generates the Go type name for the custom resource list
func (info *CRDInfo) GetListTypeName() string {
	return info.ListKind
}

// GetAPIVersion returns the full API version string (group/version)
func (info *CRDInfo) GetAPIVersion() string {
	if info.Group == "" {
		return info.Version
	}
	return fmt.Sprintf("%s/%s", info.Group, info.Version)
}

// HasShortNames returns true if the CRD defines short names
func (info *CRDInfo) HasShortNames() bool {
	return len(info.ShortNames) > 0
}

// GetGroupVersionKind returns the full GroupVersionKind string
func (info *CRDInfo) GetGroupVersionKind() string {
	return fmt.Sprintf("%s/%s, Kind=%s", info.Group, info.Version, info.Kind)
}

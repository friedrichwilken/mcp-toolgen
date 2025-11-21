package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// GoTypeInfo represents information about a Go type to be generated
type GoTypeInfo struct {
	Name        string                 // Go type name
	JSONName    string                 // JSON field name
	GoType      string                 // Go type string (e.g., "string", "*int32", "[]MyType")
	JSONTag     string                 // Complete JSON tag
	Description string                 // Field description/comment
	Required    bool                   // Whether the field is required
	Properties  map[string]*GoTypeInfo // For object types, nested properties
	Items       *GoTypeInfo            // For array types, the item type
}

// SchemaAnalyzer analyzes OpenAPI v3 schemas and generates Go type information
type SchemaAnalyzer struct {
	typeCache map[string]*GoTypeInfo
}

// NewSchemaAnalyzer creates a new SchemaAnalyzer
func NewSchemaAnalyzer() *SchemaAnalyzer {
	return &SchemaAnalyzer{
		typeCache: make(map[string]*GoTypeInfo),
	}
}

// AnalyzeSchema analyzes an OpenAPI v3 schema and returns Go type information
func (s *SchemaAnalyzer) AnalyzeSchema(schema *apiextensionsv1.JSONSchemaProps, typeName, fieldName string) (*GoTypeInfo, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", typeName, fieldName)
	if cached, exists := s.typeCache[cacheKey]; exists {
		return cached, nil
	}

	typeInfo := &GoTypeInfo{
		Name:        typeName,
		JSONName:    fieldName,
		Description: schema.Description,
	}

	// Determine Go type based on schema type
	goType, err := s.getGoTypeFromSchema(schema, typeName)
	if err != nil {
		return nil, fmt.Errorf("failed to determine Go type for %s: %w", typeName, err)
	}
	typeInfo.GoType = goType

	// Generate JSON tag
	typeInfo.JSONTag = s.generateJSONTag(fieldName, schema)

	// Handle object types with properties
	if schema.Type == "object" && len(schema.Properties) > 0 {
		typeInfo.Properties = make(map[string]*GoTypeInfo)
		requiredFields := make(map[string]bool)
		for _, required := range schema.Required {
			requiredFields[required] = true
		}

		for propName := range schema.Properties {
			propSchema := schema.Properties[propName]
			propTypeName := s.generatePropertyTypeName(typeName, propName)
			propInfo, err := s.AnalyzeSchema(&propSchema, propTypeName, propName)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze property %s: %w", propName, err)
			}
			propInfo.Required = requiredFields[propName]
			typeInfo.Properties[propName] = propInfo
		}
	}

	// Handle array types
	if schema.Type == "array" && schema.Items != nil && schema.Items.Schema != nil {
		itemTypeName := s.generateItemTypeName(typeName)
		itemInfo, err := s.AnalyzeSchema(schema.Items.Schema, itemTypeName, "")
		if err != nil {
			return nil, fmt.Errorf("failed to analyze array items for %s: %w", typeName, err)
		}
		typeInfo.Items = itemInfo
	}

	// Cache the result
	s.typeCache[cacheKey] = typeInfo
	return typeInfo, nil
}

const (
	goTypeString = "string"
)

// getGoTypeFromSchema determines the appropriate Go type for a given schema
//
//nolint:gocyclo // Complex type mapping logic is necessary for comprehensive CRD schema support
func (s *SchemaAnalyzer) getGoTypeFromSchema(schema *apiextensionsv1.JSONSchemaProps, typeName string) (string, error) {
	switch schema.Type {
	case goTypeString:
		if len(schema.Enum) > 0 {
			// For enums, we could generate a custom type, but for simplicity use string
			return goTypeString, nil
		}
		return goTypeString, nil

	case "integer":
		if schema.Format == "int64" {
			return "int64", nil
		}
		return "int32", nil

	case "number":
		if schema.Format == "double" {
			return "float64", nil
		}
		return "float32", nil

	case "boolean":
		return "bool", nil

	case "array":
		if schema.Items != nil && schema.Items.Schema != nil {
			itemTypeName := s.generateItemTypeName(typeName)
			itemType, err := s.getGoTypeFromSchema(schema.Items.Schema, itemTypeName)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("[]%s", itemType), nil
		}
		return "[]interface{}", nil

	case "object":
		if len(schema.Properties) > 0 {
			// This is a structured object, use the type name
			return typeName, nil
		}
		// Generic object
		return "map[string]interface{}", nil

	case "":
		// No type specified, check for other indicators
		if len(schema.Properties) > 0 {
			return typeName, nil
		}
		if schema.Items != nil {
			itemTypeName := s.generateItemTypeName(typeName)
			itemType, err := s.getGoTypeFromSchema(schema.Items.Schema, itemTypeName)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("[]%s", itemType), nil
		}
		return "interface{}", nil

	default:
		return "interface{}", nil
	}
}

// generateJSONTag creates the appropriate JSON tag for a field
func (s *SchemaAnalyzer) generateJSONTag(fieldName string, schema *apiextensionsv1.JSONSchemaProps) string {
	// Add omitempty for optional fields
	if !s.isRequired(schema) {
		return fmt.Sprintf(`json:%q`, fieldName+",omitempty")
	}
	return fmt.Sprintf(`json:%q`, fieldName)
}

// isRequired determines if a field should be marked as required (not omitempty)
func (s *SchemaAnalyzer) isRequired(schema *apiextensionsv1.JSONSchemaProps) bool {
	// This is context-dependent and would be set by the caller
	// For now, we assume optional unless explicitly marked
	return false
}

// generatePropertyTypeName creates a Go type name for a nested property
func (s *SchemaAnalyzer) generatePropertyTypeName(parentType, propName string) string {
	// Convert property name to Go naming conventions
	goName := s.toGoName(propName)
	return fmt.Sprintf("%s%s", parentType, goName)
}

// generateItemTypeName creates a Go type name for array items
func (s *SchemaAnalyzer) generateItemTypeName(arrayType string) string {
	// Remove trailing 's' if present and add 'Item'
	if strings.HasSuffix(arrayType, "s") {
		return arrayType[:len(arrayType)-1] + "Item"
	}
	return arrayType + "Item"
}

// toGoName converts a JSON field name to Go naming conventions
func (s *SchemaAnalyzer) toGoName(name string) string {
	// Split on common separators and capitalize each part
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})

	var result []string
	caser := cases.Title(language.English)
	for _, part := range parts {
		if part != "" {
			result = append(result, caser.String(strings.ToLower(part)))
		}
	}

	return strings.Join(result, "")
}

// GetStructFields returns all struct fields for a type, sorted by name
func (typeInfo *GoTypeInfo) GetStructFields() []*GoTypeInfo {
	if typeInfo.Properties == nil {
		return nil
	}

	var fields []*GoTypeInfo
	for _, prop := range typeInfo.Properties {
		fields = append(fields, prop)
	}

	// Sort by Go field name for consistent output
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	return fields
}

// IsComplexType returns true if this represents a complex type (struct)
func (typeInfo *GoTypeInfo) IsComplexType() bool {
	return len(typeInfo.Properties) > 0
}

// IsArrayType returns true if this represents an array type
func (typeInfo *GoTypeInfo) IsArrayType() bool {
	return strings.HasPrefix(typeInfo.GoType, "[]")
}

// IsPrimitiveType returns true if this represents a primitive Go type
func (typeInfo *GoTypeInfo) IsPrimitiveType() bool {
	primitives := map[string]bool{
		"string":  true,
		"int":     true,
		"int32":   true,
		"int64":   true,
		"float32": true,
		"float64": true,
		"bool":    true,
	}
	return primitives[typeInfo.GoType]
}

// GetGoFieldName returns the Go field name (capitalized JSON name)
func (typeInfo *GoTypeInfo) GetGoFieldName() string {
	if typeInfo.Name != "" {
		return typeInfo.Name
	}
	caser := cases.Title(language.English)
	return caser.String(typeInfo.JSONName)
}

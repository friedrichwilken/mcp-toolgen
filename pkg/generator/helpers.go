package generator

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Template helper functions

// toLower converts a string to lowercase
func toLower(s string) string {
	return strings.ToLower(s)
}

// toUpper converts a string to uppercase
func toUpper(s string) string {
	return strings.ToUpper(s)
}

// toTitle converts a string to title case
func toTitle(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Split on common separators
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || unicode.IsSpace(r)
	})

	// If only one part, check if it's camelCase/PascalCase and split it
	if len(parts) == 1 {
		parts = splitCamelCase(parts[0])
	}

	if len(parts) == 0 {
		return s
	}

	// First part stays lowercase, rest are title cased
	result := strings.ToLower(parts[0])
	caser := cases.Title(language.English)
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			result += caser.String(strings.ToLower(parts[i]))
		}
	}

	return result
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	// First split on common separators (-, _, ., space)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || unicode.IsSpace(r)
	})

	// If only one part, check if it's camelCase and split it
	if len(parts) == 1 {
		parts = splitCamelCase(parts[0])
	}

	var result []string
	caser := cases.Title(language.English)
	for _, part := range parts {
		if part != "" {
			result = append(result, caser.String(strings.ToLower(part)))
		}
	}

	return strings.Join(result, "")
}

// splitCamelCase splits a camelCase or PascalCase string into words
func splitCamelCase(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Found a capital letter, start new word
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase letters (except first)
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else if r == '-' || r == '.' || unicode.IsSpace(r) {
			// Replace separators with underscore
			result = append(result, '_')
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add hyphen before uppercase letters (except first)
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(r))
		} else if r == '_' || r == '.' || unicode.IsSpace(r) {
			// Replace separators with hyphen
			result = append(result, '-')
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// pluralize returns a simple pluralized form of a word
func pluralize(s string) string {
	if s == "" {
		return s
	}

	// Check for common words that are already plural or need special handling
	commonPlurals := map[string]string{
		"functions":   "functions",
		"services":    "services",
		"resources":   "resources",
		"policies":    "policies",
		"deployments": "deployments",
		"configmaps":  "configmaps",
		"secrets":     "secrets",
		"endpoints":   "endpoints",
		"namespaces":  "namespaces",
		"nodes":       "nodes",
		"pods":        "pods",
		"volumes":     "volumes",
		"events":      "events",
		"jobs":        "jobs",
		"cronjobs":    "cronjobs",
		"ingresses":   "ingresses",
		"classes":     "classes",
		"databases":   "databases",
		"caches":      "caches",
		"queues":      "queues",
		"processes":   "processes",
		"addresses":   "addresses",
		"responses":   "responses",
		"requests":    "requests",
		"statuses":    "statuses",
		"data":        "data",
		"metadata":    "metadata",
		"widgets":     "widgets",
	}

	// Check if word is in our known plurals map
	lowerS := strings.ToLower(s)
	if plural, exists := commonPlurals[lowerS]; exists {
		// Preserve original case
		if s == strings.ToUpper(s) {
			return strings.ToUpper(plural)
		} else if len(s) > 0 && unicode.IsUpper(rune(s[0])) {
			caser := cases.Title(language.English)
			return caser.String(plural)
		}
		return plural
	}

	// Simple pluralization rules for other words
	switch {
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"):
		return s + "es"
	case strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	case strings.HasSuffix(s, "y"):
		if len(s) > 1 && isVowel(rune(s[len(s)-2])) {
			return s + "s"
		}
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "f"):
		return s[:len(s)-1] + "ves"
	case strings.HasSuffix(s, "fe"):
		return s[:len(s)-2] + "ves"
	default:
		return s + "s"
	}
}

// singularize returns a simple singularized form of a word
func singularize(s string) string {
	if s == "" {
		return s
	}

	// Simple singularization rules
	switch {
	case strings.HasSuffix(s, "ies"):
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(s, "ves"):
		if strings.HasSuffix(s, "ives") {
			return s[:len(s)-3] + "fe"
		}
		return s[:len(s)-3] + "f"
	case strings.HasSuffix(s, "ses"), strings.HasSuffix(s, "xes"), strings.HasSuffix(s, "zes"):
		return s[:len(s)-2]
	case strings.HasSuffix(s, "ches"), strings.HasSuffix(s, "shes"):
		return s[:len(s)-2]
	case strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss"):
		return s[:len(s)-1]
	default:
		return s
	}
}

// isVowel checks if a rune is a vowel
func isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// join joins a slice of strings with a separator
func join(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

// quote wraps a string in quotes
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

// indent adds indentation to each line of a string
func indent(s, indentStr string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, indentStr+line)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// wrapComment wraps text for Go comments
func wrapComment(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	var currentLine strings.Builder
	currentLine.WriteString("// ")

	for _, word := range words {
		// Check if adding this word would exceed the limit
		if currentLine.Len()+len(word)+1 > maxWidth {
			// Start a new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString("// ")
		}

		if currentLine.Len() > 3 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add the last line
	if currentLine.Len() > 3 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// escapeString escapes a string for use in Go string literals
// It handles newlines, quotes, and other special characters
func escapeString(s string) string {
	// Replace backslashes first to avoid double-escaping
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Replace quotes
	s = strings.ReplaceAll(s, `"`, `\"`)
	// Replace newlines with space to keep descriptions on one line
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	// Replace tabs with spaces
	s = strings.ReplaceAll(s, "\t", " ")
	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

// generateFieldName generates a Go field name from a JSON field name
func generateFieldName(jsonName string) string {
	return toPascalCase(jsonName)
}

// generateMethodName generates a Go method name
func generateMethodName(operation, resourceName string) string {
	switch operation {
	case "create":
		return fmt.Sprintf("Create%s", toPascalCase(resourceName))
	case "get":
		return fmt.Sprintf("Get%s", toPascalCase(resourceName))
	case "list":
		return fmt.Sprintf("List%s", toPascalCase(pluralize(resourceName)))
	case "update":
		return fmt.Sprintf("Update%s", toPascalCase(resourceName))
	case "delete":
		return fmt.Sprintf("Delete%s", toPascalCase(resourceName))
	default:
		return toPascalCase(operation + resourceName)
	}
}

// generateToolName generates an MCP tool name
func generateToolName(operation, resourceName string) string {
	switch operation {
	case "create":
		return fmt.Sprintf("%s_%s", toSnakeCase(resourceName), "create")
	case "get":
		return fmt.Sprintf("%s_%s", toSnakeCase(resourceName), "get")
	case "list":
		return fmt.Sprintf("%s_%s", toSnakeCase(pluralize(resourceName)), "list")
	case "update":
		return fmt.Sprintf("%s_%s", toSnakeCase(resourceName), "update")
	case "delete":
		return fmt.Sprintf("%s_%s", toSnakeCase(resourceName), "delete")
	default:
		return fmt.Sprintf("%s_%s", toSnakeCase(resourceName), toSnakeCase(operation))
	}
}

// convertSchemaToGoCode converts an OpenAPI schema to Go code that generates a JSON schema
// This is used in templates to generate schema definitions
// Accepts both pointer and value types - if value is passed, takes its address
func convertSchemaToGoCode(schemaInterface interface{}, indent int) string {
	schema := normalizeSchemaInterface(schemaInterface)
	if schema == nil {
		return ""
	}

	indentStr := strings.Repeat("\t", indent)
	var sb strings.Builder

	sb.WriteString("&jsonschema.Schema{\n")
	appendBasicSchemaFields(&sb, schema, indentStr)
	appendSchemaValidation(&sb, schema, indentStr)
	appendSchemaStructure(&sb, schema, indentStr, indent)
	sb.WriteString(fmt.Sprintf("%s}", indentStr))

	return sb.String()
}

// normalizeSchemaInterface handles both pointer and value types for schema
func normalizeSchemaInterface(schemaInterface interface{}) *apiextensionsv1.JSONSchemaProps {
	switch v := schemaInterface.(type) {
	case *apiextensionsv1.JSONSchemaProps:
		return v
	case apiextensionsv1.JSONSchemaProps:
		return &v
	default:
		return nil
	}
}

// appendBasicSchemaFields appends type, description, and enum to schema code
func appendBasicSchemaFields(sb *strings.Builder, schema *apiextensionsv1.JSONSchemaProps, indentStr string) {
	if schema.Type != "" {
		fmt.Fprintf(sb, "%s\tType:        %q,\n", indentStr, schema.Type)
	}

	if schema.Description != "" {
		desc := strings.ReplaceAll(schema.Description, `"`, `\"`)
		fmt.Fprintf(sb, "%s\tDescription: %q,\n", indentStr, desc)
	}

	if len(schema.Enum) > 0 {
		fmt.Fprintf(sb, "%s\tEnum:        []any{", indentStr)
		for i, val := range schema.Enum {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(val.Raw))
		}
		sb.WriteString("},\n")
	}
}

// appendSchemaValidation appends validation constraints to schema code
func appendSchemaValidation(sb *strings.Builder, schema *apiextensionsv1.JSONSchemaProps, indentStr string) {
	if schema.Minimum != nil {
		fmt.Fprintf(sb, "%s\tMinimum:     %v,\n", indentStr, *schema.Minimum)
	}
	if schema.Maximum != nil {
		fmt.Fprintf(sb, "%s\tMaximum:     %v,\n", indentStr, *schema.Maximum)
	}
	if schema.MinLength != nil {
		fmt.Fprintf(sb, "%s\tMinLength:   %v,\n", indentStr, *schema.MinLength)
	}
	if schema.MaxLength != nil {
		fmt.Fprintf(sb, "%s\tMaxLength:   %v,\n", indentStr, *schema.MaxLength)
	}
	if schema.Pattern != "" {
		fmt.Fprintf(sb, "%s\tPattern:     %q,\n", indentStr, schema.Pattern)
	}
}

// appendSchemaStructure appends properties, required fields, items, and additional properties
func appendSchemaStructure(sb *strings.Builder, schema *apiextensionsv1.JSONSchemaProps, indentStr string, indent int) {
	if len(schema.Properties) > 0 {
		fmt.Fprintf(sb, "%s\tProperties: map[string]*jsonschema.Schema{\n", indentStr)
		for propName := range schema.Properties {
			propSchema := schema.Properties[propName]
			fmt.Fprintf(sb, "%s\t\t%q: ", indentStr, propName)
			sb.WriteString(convertSchemaToGoCode(&propSchema, indent+2))
			sb.WriteString(",\n")
		}
		fmt.Fprintf(sb, "%s\t},\n", indentStr)
	}

	if len(schema.Required) > 0 {
		fmt.Fprintf(sb, "%s\tRequired:    []string{", indentStr)
		for i, req := range schema.Required {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%q", req)
		}
		sb.WriteString("},\n")
	}

	if schema.Items != nil && schema.Items.Schema != nil {
		fmt.Fprintf(sb, "%s\tItems: &jsonschema.Schema", indentStr)
		sb.WriteString(convertSchemaToGoCode(schema.Items.Schema, indent+1))
		sb.WriteString(",\n")
	}

	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		fmt.Fprintf(sb, "%s\tAdditionalProperties: &jsonschema.Schema", indentStr)
		sb.WriteString(convertSchemaToGoCode(schema.AdditionalProperties.Schema, indent+1))
		sb.WriteString(",\n")
	}
}

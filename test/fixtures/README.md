# Test Fixtures

This directory contains test CRD files for comprehensive testing of mcp-toolgen.

## Test Cases

### simple-crd.yaml
- **Purpose**: Basic CRD with simple types
- **Features**: string, integer, boolean fields with validation
- **Scope**: Namespaced
- **Kind**: Widget
- **Use**: Unit tests for basic type generation

### complex-crd.yaml
- **Purpose**: Complex CRD with nested objects and arrays
- **Features**:
  - Nested objects (selector, template, resources)
  - Arrays with complex item schemas
  - Enums and defaults
  - Deep property nesting
- **Scope**: Namespaced
- **Kind**: Application
- **Use**: Integration tests for complex type structures

### cluster-scoped-crd.yaml
- **Purpose**: Cluster-scoped resource (not namespaced)
- **Features**:
  - Cluster scope
  - Pattern validation (regex)
  - Additional properties
  - Complex enum patterns
- **Scope**: Cluster
- **Kind**: GlobalConfig
- **Use**: Testing cluster-scoped resource generation

### multi-version-crd.yaml
- **Purpose**: Multi-version CRD with storage version
- **Features**:
  - Multiple API versions (v1alpha1, v1beta1, v1)
  - Storage version detection
  - Version evolution patterns
  - Complex validation rules
- **Scope**: Namespaced
- **Kind**: Database
- **Use**: Testing multi-version CRD handling

## Usage in Tests

These fixtures can be used in:

1. **Unit Tests**: Test individual analyzer components
2. **Integration Tests**: Test end-to-end generation workflow
3. **Validation Tests**: Verify generated code correctness
4. **Regression Tests**: Ensure changes don't break existing functionality

## File Sizes

- simple-crd.yaml: ~800 bytes (basic)
- complex-crd.yaml: ~5.2KB (moderate complexity)
- cluster-scoped-crd.yaml: ~2.8KB (cluster scope features)
- multi-version-crd.yaml: ~5.8KB (version complexity)

Total: ~15KB of comprehensive test data
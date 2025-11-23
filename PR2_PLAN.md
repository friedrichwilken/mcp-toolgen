# PR #2: Fix handlers.go.tmpl and schema.go.tmpl API Compatibility

## Status
- **Branch**: `fix/handlers-schema-api-compatibility`
- **Created**: 2025-11-22
- **Status**: In Progress

## Problem Statement

The current handlers.go.tmpl and schema.go.tmpl templates generate code that's incompatible with the kubernetes-mcp-server API:

### handlers.go.tmpl Issues
1. **Line 44, 105, 147, 189, 250**: Calls `params.GetClient()` which doesn't exist
   - `ToolHandlerParams` embeds `*internalk8s.Kubernetes` but doesn't have a `GetClient()` method
2. **Lines 52-86**: Uses custom client approach with `New{Kind}Client()` pattern
   - Not compatible with k8sms architecture
   - Should use generic `params.Resources*()` methods instead
3. **Lines 290-300**: Custom `parseResource()` helper not aligned with k8sms patterns
   - k8sms uses YAML marshaling for resources

### schema.go.tmpl Issues
1. **Line 63, 68**: Type mismatch - using `1` (int) instead of `ptr.To(float64(1))` for Minimum/Maximum
2. **Line 123**: Type mismatch - using `1` (int) instead of `ptr.To(1)` for MinLength
3. **Line 76, 133, 140**: Syntax errors with `jsonschema.Schema` type usage
   - Appears to be missing `&` operator for pointer types

## Solution Architecture

### handlers.go.tmpl Rewrite

The handlers should follow the pattern from `kubernetes-mcp-server/pkg/toolsets/core/resources.go`:

**Key Changes**:
1. Remove dependency on controller-runtime client
2. Use `params.Resources*()` methods from embedded Kubernetes struct
3. Construct `schema.GroupVersionKind` from CRD metadata
4. Use YAML/JSON for resource serialization

**Pattern to Follow**:
```go
func handle{Kind}Get(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
    args := params.GetArguments()

    // Extract parameters
    namespace := args["namespace"]
    if namespace == nil {
        namespace = ""
    }
    name := args["name"]
    if name == nil {
        return api.NewToolCallResult("", errors.New("name is required")), nil
    }

    // Construct GVK
    gvk := &schema.GroupVersionKind{
        Group:   "{CRD.Group}",
        Version: "{CRD.Version}",
        Kind:    "{CRD.Kind}",
    }

    // Call generic resource method
    ns, _ := namespace.(string)
    n, _ := name.(string)
    ret, err := params.ResourcesGet(params, gvk, ns, n)
    if err != nil {
        return api.NewToolCallResult("", fmt.Errorf("failed to get {kind}: %v", err)), nil
    }

    return api.NewToolCallResult(output.MarshalYaml(ret)), nil
}
```

**Operations Mapping**:
- **create** → `params.ResourcesCreateOrUpdate(params, yamlString)`
- **get** → `params.ResourcesGet(params, gvk, namespace, name)`
- **list** → `params.ResourcesList(params, gvk, namespace, options)`
- **update** → `params.ResourcesCreateOrUpdate(params, yamlString)`
- **delete** → `params.ResourcesDelete(params, gvk, namespace, name)`

**Required Imports**:
```go
import (
    "errors"
    "fmt"

    "github.com/containers/kubernetes-mcp-server/pkg/api"
    internalk8s "github.com/containers/kubernetes-mcp-server/pkg/kubernetes"
    "github.com/containers/kubernetes-mcp-server/pkg/output"
    "k8s.io/apimachinery/pkg/runtime/schema"
)
```

### schema.go.tmpl Fixes

**Issue 1: Numeric Type Mismatches**

Current (broken):
```go
Minimum: 1,
Maximum: 1,
MinLength: 1,
```

Fixed:
```go
Minimum: ptr.To(float64(1)),
Maximum: ptr.To(float64(1)),
MinLength: ptr.To(int64(1)),
```

**Issue 2: Schema Pointer Syntax**

Current (broken):
```go
Items: jsonschema.Schema{Type: "object"},
```

Fixed:
```go
Items: &jsonschema.Schema{Type: "object"},
```

**Required Import**:
```go
import (
    "github.com/google/jsonschema-go/jsonschema"
    "k8s.io/utils/ptr"
)
```

## Implementation Plan

### Step 1: Rewrite handlers.go.tmpl
1. ✅ Research k8sms Resources API patterns
2. Create simplified handler functions using `params.Resources*()` methods
3. Update imports to match k8sms patterns
4. Remove custom client code
5. Test with Kyma Functions CRD

### Step 2: Fix schema.go.tmpl
1. Add `ptr` package import
2. Fix numeric field types (Minimum, Maximum, MinLength, MaxLength)
3. Fix schema pointer syntax for Items and AdditionalProperties
4. Ensure all numeric constraints use proper types

### Step 3: Update client.go.tmpl (if needed)
- May need to remove or simplify since handlers won't use it
- Or keep as utility for advanced use cases

### Step 4: Test Generated Code
1. Generate Functions toolset: `./build/mcp-toolgen --crd crds/kyma-serverless-function-crd.yaml --output /tmp/functions-test --package functions --module-path test`
2. Verify compilation: `cd /tmp/functions-test && go mod init test && go mod tidy && go build`
3. Check for remaining errors
4. Update golden files if needed

### Step 5: Integration Test with ek8sms
1. Generate Functions toolset in ek8sms: `--output extendable-kubernetes-mcp-server/pkg/functions`
2. Enable import in ek8sms modules.go
3. Build ek8sms: `cd extendable-kubernetes-mcp-server && make build`
4. Run basic MCP protocol tests

## Template Changes Summary

### handlers.go.tmpl - Before (320 lines)
- Complex client-based approach
- Custom client wrapper pattern
- 5 handler functions with custom logic
- Helper functions for parsing

### handlers.go.tmpl - After (estimated ~200 lines)
- Simple Resources API approach
- Direct use of params.Resources* methods
- 5 handler functions with unified pattern
- Minimal helper functions

### schema.go.tmpl - Before
```go
Type: "integer",
Minimum: 1,
Maximum: 100,
Items: jsonschema.Schema{Type: "object"},
```

### schema.go.tmpl - After
```go
Type: "integer",
Minimum: ptr.To(float64(1)),
Maximum: ptr.To(float64(100)),
Items: &jsonschema.Schema{Type: "object"},
```

## Testing Strategy

### Unit Tests
- Existing unit tests should continue to pass
- May need to update generator tests for new patterns

### Integration Tests
- Generate test CRDs (simple-crd, cluster-scoped-crd)
- Verify all generated files compile
- Check golden file differences

### E2E Tests
- Generate Functions toolset
- Build with ek8sms
- Test basic CRUD operations via MCP protocol

## Expected Outcomes

### After PR #2 Merge
1. ✅ Generated handlers.go compiles without errors
2. ✅ Generated schema.go compiles without errors
3. ✅ Generated code integrates cleanly with ek8sms
4. ✅ Functions toolset can be enabled in ek8sms
5. ✅ All CRUD operations work via MCP protocol

### Generated Code Quality
- Clean, idiomatic Go code
- Follows k8sms architectural patterns
- Minimal dependencies
- Easy to understand and maintain

## Risk Assessment

### Low Risk
- schema.go.tmpl fixes are straightforward type corrections
- Well-understood API patterns from k8sms

### Medium Risk
- handlers.go.tmpl is a significant rewrite
- Need to ensure all CRUD operations work correctly
- Golden files will need updates

### Mitigation
- Incremental testing at each step
- Reference k8sms implementations closely
- Keep PR focused on API compatibility only

## Timeline Estimate

- handlers.go.tmpl rewrite: 1-2 hours
- schema.go.tmpl fixes: 30 minutes
- Testing and validation: 1 hour
- Golden file updates: 30 minutes
- PR creation and review: 30 minutes

**Total**: 3-4 hours of focused work

## Success Criteria

- [ ] Generated handlers.go compiles
- [ ] Generated schema.go compiles
- [ ] Generated types.go compiles (already ✅ from PR #27)
- [ ] All tests pass
- [ ] Golden files updated
- [ ] Functions toolset integrates with ek8sms
- [ ] ek8sms builds successfully
- [ ] Basic CRUD operations tested

## Follow-up Work (Future PRs)

After PR #2, potential improvements:
- Add better error handling in generated handlers
- Add support for label selectors in list operations
- Add support for field selectors
- Add pagination support for large lists
- Add watch/stream support for resources
- Generate OpenAPI schemas for better validation
- Add metrics/logging to generated handlers

## References

- kubernetes-mcp-server/pkg/toolsets/core/resources.go - Generic resource handlers
- kubernetes-mcp-server/pkg/toolsets/core/pods.go - Specific resource handlers
- kubernetes-mcp-server/pkg/api/toolsets.go - ToolHandlerParams definition
- kubernetes-mcp-server/pkg/kubernetes/resources.go - Resources API methods

## Notes

- This PR is separate from PR #27 (nested type generation) by design
- Keeps changes focused and reviewable
- Each PR addresses a distinct concern
- PR #27 must merge before PR #2 to avoid conflicts

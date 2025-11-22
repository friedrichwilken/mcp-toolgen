# Golden Files for Template Testing

This directory contains "golden files" - expected outputs from code generation that are used to verify that template changes don't unexpectedly alter generated code.

## Purpose

Golden files serve as regression tests for the code generation templates. When templates are modified:
1. Tests compare generated output against golden files
2. Any differences indicate potential breaking changes
3. Intentional changes can be approved by updating golden files

## Structure

Each subdirectory contains golden files for a specific test case:

- `simple_crd_with_all_operations/` - Simple CRD with all CRUD operations (create, get, list, update, delete)
- `simple_crd_with_create_and_read_only/` - Simple CRD with only create, get, and list operations
- `cluster_scoped_crd/` - Cluster-scoped CRD (not namespace-scoped)

Each test case directory contains the complete set of generated files:
- `toolset.go` - MCP toolset registration and tool definitions
- `types.go` - Go types matching CRD schema
- `client.go` - Kubernetes client wrapper
- `handlers.go` - MCP tool handlers
- `schema.go` - JSON schemas for validation
- `doc.go` - Package documentation

## Running Tests

### Normal Test Run
```bash
# Run golden file tests
go test ./test/integration/ -run TestTemplateGoldenFiles

# Run all integration tests
go test ./test/integration/
```

### Updating Golden Files

When templates are intentionally changed and you want to update the golden files:

```bash
# Update all golden files
go test ./test/integration/ -run TestTemplateGoldenFiles -update-golden

# Review the changes
git diff test/fixtures/golden/

# If changes are expected, commit them
git add test/fixtures/golden/
git commit -m "Update golden files for template changes"
```

## Adding New Test Cases

To add a new test case:

1. **Add test case to `template_golden_test.go`**:
   ```go
   {
       name:        "my new test case",
       crdFile:     "my-crd.yaml",
       packageName: "mypackage",
       operations:  []string{"create", "get", "list"},
       expectedFiles: []string{
           "toolset.go",
           "types.go",
           "client.go",
           "handlers.go",
           "schema.go",
           "doc.go",
       },
       validateFunc: validateMyTestCase,
   },
   ```

2. **Create validation function (optional)**:
   ```go
   func validateMyTestCase(t *testing.T, goldenDir, generatedDir string) {
       // Add custom validation logic
   }
   ```

3. **Generate initial golden files**:
   ```bash
   go test ./test/integration/ -run TestTemplateGoldenFiles -update-golden
   ```

4. **Review and commit**:
   ```bash
   git add test/fixtures/golden/my_new_test_case/
   git commit -m "Add golden files for my new test case"
   ```

## Best Practices

1. **Keep golden files minimal** - Use the simplest CRDs that test the specific scenario
2. **Review changes carefully** - When updating golden files, review diffs to ensure changes are intentional
3. **Test before committing** - Always run tests without `-update-golden` before committing
4. **Document test cases** - Add comments explaining what each test case validates
5. **Version control** - Always commit golden files with the corresponding template changes

## Troubleshooting

### Test failures after template changes
1. Review the diff shown in test output
2. If changes are expected, run with `-update-golden`
3. Review the updated golden files with `git diff`
4. Commit if changes are correct

### Golden file missing error
```bash
# Regenerate missing golden files
go test ./test/integration/ -run TestTemplateGoldenFiles -update-golden
```

### Golden file mismatch
The test output shows the first 10 differing lines. To see full diff:
```bash
# Compare manually
diff -u test/fixtures/golden/test_case/file.go /tmp/generated/file.go
```

## Integration with CI/CD

Golden file tests run automatically in CI/CD:
- Tests fail if generated code doesn't match golden files
- Prevents accidental template changes from being merged
- Forces explicit review of template changes through golden file updates

## Related Files

- `test/integration/template_golden_test.go` - Golden file test implementation
- `test/fixtures/*.yaml` - CRD fixtures used for generation
- `pkg/generator/templates/*.tmpl` - Code generation templates

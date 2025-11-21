# CLAUDE PROJECT CONTEXT: MCP Toolgen

## PROJECT_OVERVIEW

**Name**: mcp-toolgen
**Purpose**: Code generation tool that creates Go toolsets from Kubernetes Custom Resource Definitions (CRDs)
**Status**: Active development
**Go Version**: 1.25+
**Module Path**: github.com/friedrichwilken/mcp-toolgen

## CORE_CONCEPT

The mcp-toolgen bridges the gap between Custom Resource Definitions and the extendable-kubernetes-mcp-server by:

1. **INPUT**: CRD YAML files (e.g., `kyma-serverless-function-crd.yaml`)
2. **PROCESS**: Analyzes CRD structure, schemas, and metadata
3. **OUTPUT**: Complete Go packages with MCP toolsets for CRUD operations
4. **INTEGRATION**: Generated packages seamlessly integrate with ek8sms

## DIRECTORY_STRUCTURE

```
mcp-toolgen/
├── CLAUDE.md                          # This context file
├── README.md                          # User documentation
├── go.mod                             # Go module definition (Go 1.25)
├── go.sum                             # Dependency checksums
├── .golangci.yml                      # Linter configuration
├── Makefile                           # Build automation
│
├── cmd/
│   └── mcp-toolgen/
│       └── main.go                    # CLI entry point
│
├── pkg/
│   ├── analyzer/                      # CRD parsing & analysis
│   │   ├── crd.go                     # CRD YAML parsing
│   │   ├── crd_test.go                # CRD parsing tests
│   │   ├── schema.go                  # OpenAPI v3 schema analysis
│   │   ├── schema_test.go             # Schema analysis tests
│   │   └── types.go                   # Generation configuration types
│   │
│   ├── generator/                     # Code generation engine
│   │   ├── generator.go               # Main generation logic
│   │   ├── generator_test.go          # Generator tests
│   │   ├── helpers.go                 # Template helper functions
│   │   ├── writer.go                  # File output handling
│   │   ├── templates.go               # Template loading
│   │   └── templates/                 # Go code templates
│   │       ├── toolset.go.tmpl        # MCP toolset registration
│   │       ├── types.go.tmpl          # CRD Go types
│   │       ├── client.go.tmpl         # Kubernetes client wrapper
│   │       ├── handlers.go.tmpl       # MCP tool handlers
│   │       ├── schema.go.tmpl         # JSON schemas
│   │       └── doc.go.tmpl            # Package documentation
│   │
│   └── cmd/                           # CLI implementation
│       ├── root.go                    # Main CLI commands
│       └── version.go                 # Version command
│
├── test/
│   ├── fixtures/                      # Test CRD files
│   └── integration/
│       ├── README.md                  # Integration test status
│       └── e2e_test.go.disabled       # Disabled tests (need API updates)
│
├── build/                             # Build artifacts directory
│   └── mcp-toolgen                   # Compiled binary
│
└── examples/                          # Generated example toolsets
```

## KEY_COMMANDS

### Build & Installation
```bash
make build                 # Build the binary
make install               # Install to GOPATH/bin
make clean                 # Remove build artifacts
```

### Code Quality
```bash
make code-quality          # Run all quality checks (tidy, fmt, fmt-modern, lint-fix)
make fmt                   # Standard Go formatting
make fmt-modern            # Modern Go formatting (interface{} -> any)
make lint                  # Run golangci-lint
make lint-fix              # Run golangci-lint with auto-fixes
make security-check        # Run gosec + govulncheck
make pre-commit-check      # Run code-quality + tests
```

### Testing
```bash
make test                  # Run all tests
make test-unit             # Run unit tests only
make test-integration      # Run integration tests (currently disabled)
make test-coverage         # Generate coverage report
```

### Development Workflow
```bash
make dev-build             # Quick build without deps
make dev-test              # Quick test run
make check-deps            # Check for dependency updates
make qa                    # Run all QA checks
make ci                    # Run CI pipeline
```

### Generation Examples
```bash
make demo-cr               # Generate with create+read operations only
make demo-d                # Generate with delete operation only
make demo-all              # Generate with all CRUD operations
make demo-dry-run          # Dry run (no files created)
make validate-crud         # Validate CRUD flag functionality
make generate-example      # Generate for ek8sms integration
make clean-examples        # Clean generated examples
```

## DEPENDENCIES

### Main Dependencies
- **github.com/spf13/cobra v1.10.1** - CLI framework
- **github.com/spf13/viper v1.21.0** - Configuration management
- **github.com/stretchr/testify v1.11.1** - Testing framework
- **k8s.io/apiextensions-apiserver v0.34.2** - CRD types and schemas
- **k8s.io/apimachinery v0.34.2** - Kubernetes API machinery
- **sigs.k8s.io/yaml v1.6.0** - YAML parsing

### Development Tools
- **golangci-lint v2.6.2** - Comprehensive Go linter
- **gofumpt** - Strict Go formatter
- **gosec** - Security vulnerability scanner
- **govulncheck** - Go vulnerability checker

## ARCHITECTURE

### 1. CRD Analysis (pkg/analyzer/)

**CRDAnalyzer** (`crd.go`):
- Parses CRD YAML files
- Extracts metadata (group, version, kind, names)
- Validates CRD structure
- Creates `CRDInfo` and `ToolsetInfo` structures

**SchemaAnalyzer** (`schema.go`):
- Analyzes OpenAPI v3 schemas from CRDs
- Converts schema types to Go types
- Handles nested structures and arrays
- Generates type information for templates

**Key Types** (`types.go`):
- `CRDInfo`: Parsed CRD metadata
- `ToolsetInfo`: Complete information for code generation
- `TypeInfo`: Go type information from schemas
- `FieldInfo`: Field metadata including JSON tags

### 2. Code Generation (pkg/generator/)

**Generator** (`generator.go`):
- Main generation orchestrator
- Loads and executes templates
- Manages file writing
- Validates configuration

**Template System** (`templates/`):
- `toolset.go.tmpl`: MCP toolset registration, tool definitions
- `types.go.tmpl`: Go structs matching CRD schemas
- `client.go.tmpl`: Kubernetes client wrappers
- `handlers.go.tmpl`: MCP tool handlers with validation
- `schema.go.tmpl`: JSON schemas for MCP tools
- `doc.go.tmpl`: Package documentation

**Helper Functions** (`helpers.go`):
- `toPascalCase`: Convert strings to PascalCase
- `toCamelCase`: Convert strings to camelCase
- `toSnakeCase`: Convert strings to snake_case
- `pluralize`: Simple pluralization
- String manipulation utilities

**File Writer** (`writer.go`):
- Handles file I/O operations
- Optional code formatting
- Validates file structure
- Manages overwrite policies

### 3. CLI Interface (pkg/cmd/)

**Root Command** (`root.go`):
- `--crd`: Path to single CRD file
- `--crd-dir`: Directory of CRD files
- `--output`: Output directory for generated code
- `--output-base`: Base directory for multiple generations
- `--package`: Go package name
- `--module-path`: Go module path
- `--crud`: CRUD operations filter (c,r,u,d)
- `--dry-run`: Preview without writing files
- `--overwrite`: Overwrite existing files
- `--verbose`: Enable verbose logging

**Version Command** (`version.go`):
- Shows version information
- Git commit and build date

## GENERATED_CODE_STRUCTURE

For a CRD with kind `Function`, the generator creates:

```go
pkg/functions/
├── toolset.go      // MCP toolset registration
│   - FunctionToolset struct
│   - GetName() string
│   - GetTools() []mcp.Tool
│   - Handlers() map[string]mcp.Handler
│   - init() function for auto-registration
│
├── types.go        // Go types from CRD
│   - Function struct (main resource)
│   - FunctionSpec struct
│   - FunctionStatus struct
│   - FunctionList struct
│   - DeepCopy methods
│
├── client.go       // Kubernetes client wrapper
│   - FunctionClient struct
│   - Create() method
│   - Get() method
│   - List() method
│   - Update() method
│   - Delete() method
│
├── handlers.go     // MCP tool handlers
│   - handleCreateFunction()
│   - handleGetFunction()
│   - handleListFunctions()
│   - handleUpdateFunction()
│   - handleDeleteFunction()
│
├── schema.go       // JSON schemas for validation
│   - createFunctionSchema
│   - getFunctionSchema
│   - listFunctionsSchema
│   - updateFunctionSchema
│   - deleteFunctionSchema
│
└── doc.go          // Package documentation
    - Package-level documentation
    - Usage examples
```

## USAGE_EXAMPLES

### Generate from Single CRD
```bash
./build/mcp-toolgen \
  --crd ./crds/function-crd.yaml \
  --output ./pkg/functions \
  --package functions \
  --module-path github.com/myproject/server
```

### Generate with Specific Operations
```bash
# Only create and read operations
./build/mcp-toolgen --crud cr \
  --crd ./crds/function-crd.yaml \
  --output ./pkg/functions \
  --package functions \
  --module-path github.com/myproject/server
```

### Generate for ek8sms
```bash
# Direct integration with extendable-kubernetes-mcp-server
./build/mcp-toolgen \
  --crd ../crds/function-crd.yaml \
  --output ../extendable-kubernetes-mcp-server/pkg/functions \
  --package functions \
  --module-path github.com/friedrichwilken/extendable-kubernetes-mcp-server \
  --overwrite
```

### Dry Run
```bash
# Preview what would be generated
./build/mcp-toolgen --dry-run --verbose \
  --crd ./crds/function-crd.yaml \
  --output ./pkg/functions \
  --package functions \
  --module-path github.com/myproject/server
```

## CODE_QUALITY_SETUP

### Linting Configuration (.golangci.yml)

**Enabled Linters**:
- errcheck: Check error return values
- govet: Standard Go vet checks
- ineffassign: Detect ineffectual assignments
- staticcheck: Advanced static analysis
- unused: Find unused code
- bodyclose: Check HTTP response body closes
- goconst: Find repeated strings that could be constants
- gocritic: Comprehensive code analysis
- gocyclo: Cyclomatic complexity checking
- gosec: Security vulnerability detection
- lll: Line length checking (160 chars)
- misspell: Spell checking
- predeclared: Find shadowed predeclared identifiers
- unconvert: Remove unnecessary conversions
- unparam: Find unused function parameters

**Settings**:
- Line length: 160 characters
- Cyclomatic complexity: max 15
- Constant detection: min 3 chars, 3 occurrences
- Tests have relaxed rules (goconst, lll, gosec disabled)

### Modern Formatting

Uses `gofumpt` with `-extra` flag for:
- Converting `interface{}` to `any`
- Strict formatting rules
- Consistent style across codebase

### Security Scanning

- **gosec**: Scans for security vulnerabilities in Go code
- **govulncheck**: Checks dependencies for known vulnerabilities

## INTEGRATION_WITH_EK8SMS

### How Generated Code Integrates

1. **Auto-Registration**: Generated `init()` functions automatically register toolsets
2. **MCP Protocol**: Implements MCP tool interface
3. **Kubernetes Client**: Uses same client-go patterns as ek8sms
4. **Multi-Cluster**: Supports multi-cluster operations via kubeconfig

### Integration Steps

```bash
# 1. Generate toolset in ek8sms pkg/ directory
cd mcp-toolgen
make generate-example

# 2. Build ek8sms with new toolset
cd ../extendable-kubernetes-mcp-server
make build

# 3. Run ek8sms - new toolset is automatically available
./build/extendable-k8s-mcp
```

## TESTING_STATUS

### Unit Tests
- ✅ CRD parsing tests (`pkg/analyzer/crd_test.go`)
- ✅ Schema analysis tests (`pkg/analyzer/schema_test.go`)
- ⚠️  Generator tests (`pkg/generator/generator_test.go`) - partially disabled due to API changes

### Integration Tests
- ⚠️  E2E tests (`test/integration/e2e_test.go.disabled`) - needs API updates
- See `test/integration/README.md` for re-enabling instructions

### Test Coverage
Current coverage focuses on:
- CRD parsing and validation
- Schema type conversion
- Template helper functions

**TODO**: Update integration tests to work with current API

## KNOWN_ISSUES

1. **Integration Tests Disabled**: Tests reference old API methods that no longer exist
   - Need to update to use new `CRDAnalyzer` and `Generator` workflow
   - See `test/integration/README.md` for details

2. **Linter Warnings**: Some non-critical linter warnings remain:
   - Unchecked error returns in defer statements
   - Use of deprecated `strings.Title` (needs `golang.org/x/text/cases`)
   - Security warnings for file operations (expected for code generator)

3. **Test Fixtures**: May need additional test CRD files for comprehensive testing

## DEVELOPMENT_WORKFLOW

### Making Changes

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make Changes**
   - Edit code in `pkg/` directory
   - Add tests if applicable
   - Update templates in `pkg/generator/templates/` if needed

3. **Run Code Quality**
   ```bash
   make code-quality
   ```

4. **Test Changes**
   ```bash
   make test
   make demo-all  # Test with real CRD
   ```

5. **Pre-Commit Check**
   ```bash
   make pre-commit-check
   ```

6. **Build and Test Binary**
   ```bash
   make build
   ./build/mcp-toolgen --help
   ```

### Adding New Templates

1. Create template file in `pkg/generator/templates/`
2. Add template name to `files` slice in `generator.go:GenerateToolset()`
3. Update template loading in `generator.go:loadTemplates()`
4. Add helper functions in `helpers.go` if needed
5. Test with `make demo-all`

### Updating Dependencies

```bash
# Check for updates
make check-deps

# Update specific dependency
go get -u github.com/package/name@version
go mod tidy

# Or update all to latest compatible
go get -u ./...
go mod tidy
```

## RELATED_PROJECTS

- **extendable-kubernetes-mcp-server**: Target MCP server for generated code
  - Location: `../extendable-kubernetes-mcp-server/`
  - Purpose: MCP server with extensible toolsets

- **kubernetes-mcp-server**: Reference implementation
  - Location: `../kubernetes-mcp-server/`
  - Purpose: Original MCP server implementation

- **crds**: Example CRD collection
  - Location: `../crds/`
  - Contains: `kyma-serverless-function-crd.yaml` and others

## FUTURE_ENHANCEMENTS

### Planned Features
- [ ] Support for CRD webhooks in generated code
- [ ] Custom template directory support
- [ ] Batch generation from multiple CRDs
- [ ] Integration test re-enablement
- [ ] Generated code validation
- [ ] Template customization via config file
- [ ] Support for CRD conversion webhooks
- [ ] Generated documentation in multiple formats

### Code Quality Improvements
- [ ] Fix remaining linter warnings
- [ ] Improve test coverage to >80%
- [ ] Add benchmark tests for generation performance
- [ ] Document all public APIs
- [ ] Add examples directory with real-world CRDs

## TROUBLESHOOTING

### Build Failures
```bash
# Clean and rebuild
make clean
make deps
make build
```

### Template Errors
- Check template syntax in `pkg/generator/templates/`
- Verify template data structure in `generator.go:createTemplateData()`
- Test with `--dry-run --verbose` for detailed output

### Generation Errors
```bash
# Enable verbose logging
./build/mcp-toolgen --verbose --crd ./test.yaml --output ./out --package test --module-path test

# Use dry-run to see what would be generated
./build/mcp-toolgen --dry-run --verbose --crd ./test.yaml --output ./out --package test --module-path test
```

### Linter Issues
```bash
# See all issues
make lint

# Auto-fix what can be fixed
make lint-fix

# Check specific file
golangci-lint run path/to/file.go
```

## VERSION_INFORMATION

**Current Version**: Development
**Go Version**: 1.25+
**Last Updated**: 2025-11-21
**Maintainer**: friedrichwilken

## NOTES_FOR_CLAUDE

- Always use absolute paths when running commands (no `cd` due to zoxide alias)
- Use `pushd/popd` for directory changes
- Run `make code-quality` before committing changes
- Check `make help` for all available targets
- Integration tests are disabled - see `test/integration/README.md`
- Template modifications require understanding of Go text/template syntax
- Generated code must be compatible with ek8sms architecture

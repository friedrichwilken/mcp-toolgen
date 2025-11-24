# MCP Toolgen

A code generation tool that creates Go toolsets for Kubernetes Custom Resource Definitions (CRDs), specifically designed to extend the [extendable-kubernetes-mcp-server](https://github.com/friedrichwilken/extendable-kubernetes-mcp-server) with custom resource operations.

## Overview

MCP Toolgen analyzes CRD YAML files and generates complete Go packages with:
- MCP (Model Context Protocol) toolset registration
- CRUD operations for custom resources
- Kubernetes client integrations
- JSON schema validation
- Comprehensive error handling

## Features

- **CRD Analysis**: Parse and analyze CRD YAML files with OpenAPI v3 schema support
- **Code Generation**: Template-based Go code generation following established patterns
- **MCP Integration**: Generated toolsets seamlessly integrate with MCP servers
- **MCP Resource Support**: Optional CRD resource generation for LLM access to definitions
- **Multi-cluster Support**: Generated code supports multi-cluster operations
- **Type Safety**: Full Go type generation from CRD schemas
- **Flexible CRUD**: Generate specific operations (create, read, update, delete) as needed
- **Backward Compatible**: Default settings work with standard ek8sms (no resource support needed)

## Installation

```bash
# Clone the repository
git clone https://github.com/friedrichwilken/mcp-toolgen.git
cd mcp-toolgen

# Build the binary
make build

# Install to GOPATH/bin
make install
```

## Usage

### Basic Usage

```bash
# Generate toolset from a single CRD
mcp-toolgen --crd ./crds/function-crd.yaml \
            --output ./pkg/functions \
            --package functions \
            --module-path github.com/myorg/myproject

# Generate toolsets from a directory of CRDs
mcp-toolgen --crd-dir ./crds \
            --output-base ./pkg \
            --module-path github.com/myorg/myproject

# Generate with MCP resource support (requires ek8sms with resource support)
mcp-toolgen --crd ./crds/function-crd.yaml \
            --output ./pkg/functions \
            --package functions \
            --module-path github.com/myorg/myproject \
            --generate-crd-resource

# Generate with specific CRUD operations only
mcp-toolgen --crd ./crds/function-crd.yaml \
            --output ./pkg/functions \
            --package functions \
            --module-path github.com/myorg/myproject \
            --crud cr  # Only create and read operations

# Generate with custom templates
mcp-toolgen --crd ./crds/function-crd.yaml \
            --templates ./custom-templates \
            --output ./pkg/functions \
            --package functions \
            --module-path github.com/myorg/myproject
```

### Command-Line Flags

| Flag | Description | Required | Default |
|------|-------------|----------|---------|
| `--crd` | Path to a single CRD YAML file | Yes (or `--crd-dir`) | - |
| `--crd-dir` | Directory containing multiple CRD YAML files | Yes (or `--crd`) | - |
| `--output` | Output directory for generated code (single CRD) | Yes (with `--crd`) | - |
| `--output-base` | Base directory for multi-CRD generation | Yes (with `--crd-dir`) | - |
| `--package` | Go package name | No | CRD plural name |
| `--module-path` | Go module path (e.g., github.com/myorg/myproject) | Yes | - |
| `--crud` | CRUD operations to generate (c=create, r=read, u=update, d=delete) | No | `crud` (all) |
| `--generate-crd-resource` | Generate MCP resource for CRD definition | No | `false` |
| `--templates` | Custom template directory | No | embedded templates |
| `--overwrite` | Overwrite existing files | No | `false` |
| `--dry-run` | Preview generation without creating files | No | `false` |
| `--verbose` | Enable verbose logging | No | `false` |

### Integration with extendable-kubernetes-mcp-server

1. **Generate toolsets** in your ek8sms project:
   ```bash
   mcp-toolgen --crd ./crds/function-crd.yaml \
               --output ./pkg/functions \
               --package functions \
               --module-path github.com/friedrichwilken/extendable-kubernetes-mcp-server

   # Optional: Generate with MCP resource support
   mcp-toolgen --crd ./crds/function-crd.yaml \
               --output ./pkg/functions \
               --package functions \
               --module-path github.com/friedrichwilken/extendable-kubernetes-mcp-server \
               --generate-crd-resource
   ```

2. **Generated package structure**:
   ```
   pkg/functions/
   ├── toolset.go      # MCP toolset registration (+ resource support if enabled)
   ├── types.go        # Go types from CRD schema
   ├── client.go       # Kubernetes client wrapper
   ├── handlers.go     # MCP tool handlers
   ├── schema.go       # JSON schemas for validation
   └── doc.go          # Package documentation
   ```

3. **Auto-registration**: Generated toolsets automatically register with the MCP server via `init()` functions.

4. **MCP Resource Support** (optional): When `--generate-crd-resource` is enabled:
   - Generated toolset implements `ResourceProvider` interface
   - CRD YAML is embedded and exposed as an MCP resource
   - Allows LLMs to access the CRD definition directly
   - Requires ek8sms with resource support enabled

## Architecture

### Project Structure

```
mcp-toolgen/
├── cmd/mcp-toolgen/         # CLI entry point
├── pkg/
│   ├── analyzer/             # CRD parsing and analysis
│   ├── generator/            # Code generation engine
│   │   └── templates/        # Go code templates
│   └── config/               # Configuration management
├── test/
│   ├── fixtures/             # Test CRD files
│   └── integration/          # Integration tests
└── Makefile                  # Build automation
```

### Code Generation Flow

1. **Parse CRD**: Extract metadata, schema, and resource information
2. **Analyze Schema**: Convert OpenAPI v3 schema to Go type definitions
3. **Generate Code**: Apply templates to create complete Go packages
4. **Validate Output**: Ensure generated code follows patterns and compiles

## Development

### Prerequisites

- Go 1.25+
- Make

**Auto-installed tools** (installed automatically when needed):
- **goimports**: Organizes and fixes imports
- **gofumpt**: Modern Go formatter (strict, replaces `interface{}` with `any`)
- **golangci-lint**: Comprehensive linter with auto-fix capabilities
- **gosec**: Security vulnerability scanner
- **govulncheck**: Dependency vulnerability checker

### Development Setup

```bash
# Set up development environment
make dev-setup

# Run tests
make test

# AUTO-FIX ALL ISSUES (RECOMMENDED - runs all auto-fixers)
make auto-fix       # Runs: goimports → gofumpt → golangci-lint --fix

# Or run code quality checks individually
make fmt-imports    # Fix and organize imports
make fmt-modern     # Modern Go formatting (interface{} -> any)
make lint           # Check code (read-only)
make lint-fix       # Fix linting issues automatically
make security-check # Run gosec and govulncheck

# Complete workflow
make code-quality   # Run all quality checks with auto-fixes
make check          # Check without making changes (for CI)
make pre-commit-check # Quality + tests before committing
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Generate coverage report
make test-coverage
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [extendable-kubernetes-mcp-server](https://github.com/friedrichwilken/extendable-kubernetes-mcp-server) - The MCP server this tool extends
- [kubernetes-mcp-server](https://github.com/mark3labs/kubernetes-mcp-server) - Original MCP server implementation
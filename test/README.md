# MCP Toolgen Tests

This directory contains the test suite for mcp-toolgen.

## Test Structure

```
test/
├── fixtures/          # Test CRD files
│   └── testwidget-crd.yaml
├── utils/            # Test utilities
│   ├── envtest.go    # Kubernetes envtest helpers
│   └── test.go       # Common test helpers
├── e2e/              # End-to-end tests
│   └── toolgen_ek8sms_test.go
└── integration/      # (deprecated - removed)
```

## Running Tests

### Unit Tests
```bash
make test-unit
```

Tests individual components (analyzer, generator, helpers) in isolation.

### E2E Tests
```bash
# First-time setup (downloads Kubernetes binaries)
make setup-envtest

# Run E2E tests
make test-e2e
```

E2E tests validate the complete workflow:
1. Start real Kubernetes API (via envtest)
2. Apply test CRD
3. Generate toolset with mcp-toolgen
4. Build ek8sms with generated code
5. Verify the toolset works

**Requirements:**
- Go 1.25+
- ~500MB disk space for Kubernetes binaries
- ek8sms repository at `../extendable-kubernetes-mcp-server/`

### All Tests
```bash
make test
```

## E2E Test Setup

The E2E tests use [controller-runtime's envtest](https://book.kubebuilder.io/reference/envtest.html) to run a real Kubernetes API server locally.

### First Time Setup

Run the setup script to download Kubernetes binaries:

```bash
./scripts/setup-envtest.sh
```

This downloads:
- `kube-apiserver` - Kubernetes API server
- `etcd` - Key-value store for Kubernetes
- `kubectl` - Kubernetes CLI (optional)

Binaries are installed to `./test/envtest/bin/` and are ~400MB.

### Environment Variables

The setup script creates `./test/envtest/env.sh` which you can source:

```bash
source ./test/envtest/env.sh
```

Or set manually:

```bash
export KUBEBUILDER_ASSETS=/path/to/envtest/bin
```

### Troubleshooting

**"envtest not available" error:**
1. Run `make setup-envtest`
2. Check that `./test/envtest/bin/` contains `kube-apiserver` and `etcd`
3. Verify with: `./test/envtest/bin/kube-apiserver --version`

**"no space left on device" error:**
- Envtest compilation requires significant disk space
- Free up at least 1GB and try again

**Tests are slow:**
- First run downloads dependencies (~2-3 minutes)
- Subsequent runs should be faster (~30 seconds)
- Use `go test -short` to skip E2E tests

## Test Fixtures

### testwidget-crd.yaml

A simple test CRD used for E2E testing:

- **Kind**: TestWidget
- **Group**: testing.mcp-toolgen.io
- **Required fields**: `name` (string)
- **Optional fields**: `message` (string), `count` (int), `enabled` (bool)

This CRD is perfect for testing because:
- ✅ No controller required
- ✅ Tests both required and optional fields
- ✅ Multiple data types
- ✅ Simple enough to understand
- ✅ Complex enough to be realistic

## Test Utilities

### envtest.go

Provides wrappers around controller-runtime's envtest:

- `NewEnvtestEnvironment(t)` - Start Kubernetes API
- `ApplyCRDFile(t, path)` - Apply CRD to test cluster
- `CreateKubeconfigFile(t)` - Generate kubeconfig for MCP server
- `GetClient()` - Get controller-runtime client
- `Stop(t)` - Cleanup test environment

### test.go

Common test helpers:

- `TempDir(t)` - Create temporary directory
- `WriteTestFile(t, dir, name, content)` - Write test files
- `Must[T](val, err)` - Panic on error (for test setup)
- `SkipIfShort(t)` - Skip long-running tests
- `FileExists(path)`, `DirExists(path)` - File checks

## Writing New E2E Tests

Example structure:

```go
func TestMyFeature(t *testing.T) {
    utils.SkipIfShort(t)

    // 1. Start envtest
    env := utils.NewEnvtestEnvironment(t)

    // 2. Apply CRD
    crd := env.ApplyCRDFile(t, "./fixtures/my-crd.yaml")

    // 3. Generate toolset
    // ... your generation code ...

    // 4. Build and test
    // ... your test logic ...
}
```

## CI/CD

GitHub Actions workflow includes:

1. `make setup-envtest` - Install binaries
2. `make test-e2e` - Run E2E tests
3. Caches `test/envtest/` to speed up subsequent runs

## References

- [Controller Runtime Envtest](https://book.kubebuilder.io/reference/envtest.html)
- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [MCP Toolgen Documentation](../README.md)

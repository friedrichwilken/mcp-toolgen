// Package utils provides testing utilities for mcp-toolgen, including envtest setup.
package utils

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/yaml"
)

// EnvtestEnvironment wraps controller-runtime's envtest Environment
// providing a real Kubernetes API server for testing.
type EnvtestEnvironment struct {
	testEnv *envtest.Environment
	cfg     *rest.Config
	client  client.Client
	scheme  *runtime.Scheme
}

// NewEnvtestEnvironment creates and starts a new envtest environment.
// This starts a real etcd and kube-apiserver for integration testing.
func NewEnvtestEnvironment(t *testing.T) *EnvtestEnvironment {
	t.Helper()

	// Create scheme with CRD support
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s), "Failed to add core scheme")
	require.NoError(t, apiextensionsv1.AddToScheme(s), "Failed to add apiextensions scheme")

	// Setup envtest
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{}, // CRDs will be applied programmatically
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err, "Failed to start envtest")
	require.NotNil(t, cfg, "envtest config is nil")

	// Create client
	k8sClient, err := client.New(cfg, client.Options{Scheme: s})
	require.NoError(t, err, "Failed to create client")

	env := &EnvtestEnvironment{
		testEnv: testEnv,
		cfg:     cfg,
		client:  k8sClient,
		scheme:  s,
	}

	// Register cleanup
	t.Cleanup(func() {
		env.Stop(t)
	})

	return env
}

// GetConfig returns the rest.Config for the envtest cluster.
func (e *EnvtestEnvironment) GetConfig() *rest.Config {
	return e.cfg
}

// GetClient returns the controller-runtime client for the envtest cluster.
func (e *EnvtestEnvironment) GetClient() client.Client {
	return e.client
}

// GetScheme returns the runtime.Scheme used by the envtest environment.
func (e *EnvtestEnvironment) GetScheme() *runtime.Scheme {
	return e.scheme
}

// ApplyCRDFile reads a CRD from a YAML file and applies it to the cluster.
func (e *EnvtestEnvironment) ApplyCRDFile(t *testing.T, crdPath string) *apiextensionsv1.CustomResourceDefinition {
	t.Helper()

	// Read CRD file
	crdBytes := Must(ReadFileBytes(crdPath))

	// Parse CRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	require.NoError(t, yaml.Unmarshal(crdBytes, crd), "Failed to unmarshal CRD")

	// Apply CRD to cluster
	ctx := context.Background()
	require.NoError(t, e.client.Create(ctx, crd), "Failed to create CRD")

	t.Logf("Applied CRD: %s", crd.Name)
	return crd
}

// ApplyCRD applies a CRD object to the cluster.
func (e *EnvtestEnvironment) ApplyCRD(t *testing.T, crd *apiextensionsv1.CustomResourceDefinition) {
	t.Helper()

	ctx := context.Background()
	require.NoError(t, e.client.Create(ctx, crd), "Failed to create CRD")

	t.Logf("Applied CRD: %s", crd.Name)
}

// Stop stops the envtest environment and cleans up resources.
func (e *EnvtestEnvironment) Stop(t *testing.T) {
	t.Helper()

	if e.testEnv != nil {
		require.NoError(t, e.testEnv.Stop(), "Failed to stop envtest")
	}
}

// CreateKubeconfigFile creates a kubeconfig file for the envtest cluster.
// Returns the path to the generated kubeconfig file.
func (e *EnvtestEnvironment) CreateKubeconfigFile(t *testing.T) string {
	t.Helper()

	tempDir := TempDir(t)
	kubeconfigPath := filepath.Join(tempDir, "kubeconfig")

	// Create kubeconfig content
	kubeconfig := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
    certificate-authority-data: %s
  name: envtest
contexts:
- context:
    cluster: envtest
    user: envtest
  name: envtest
current-context: envtest
users:
- name: envtest
  user:
    client-certificate-data: %s
    client-key-data: %s
`,
		e.cfg.Host,
		e.cfg.CAData,
		e.cfg.CertData,
		e.cfg.KeyData,
	)

	WriteTestFile(t, tempDir, "kubeconfig", kubeconfig)
	return kubeconfigPath
}

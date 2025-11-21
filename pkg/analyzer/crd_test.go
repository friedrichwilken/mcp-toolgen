package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCRDFromFile(t *testing.T) {
	analyzer := NewCRDAnalyzer()

	tests := []struct {
		name        string
		filename    string
		wantError   bool
		wantKind    string
		wantGroup   string
		wantVersion string
		wantPlural  string
		wantScope   string
	}{
		{
			name:        "simple CRD",
			filename:    "../../test/fixtures/simple-crd.yaml",
			wantError:   false,
			wantKind:    "Widget",
			wantGroup:   "example.com",
			wantVersion: "v1",
			wantPlural:  "widgets",
			wantScope:   "Namespaced",
		},
		{
			name:        "complex CRD",
			filename:    "../../test/fixtures/complex-crd.yaml",
			wantError:   false,
			wantKind:    "Application",
			wantGroup:   "apps.example.com",
			wantVersion: "v1beta1",
			wantPlural:  "applications",
			wantScope:   "Namespaced",
		},
		{
			name:        "cluster-scoped CRD",
			filename:    "../../test/fixtures/cluster-scoped-crd.yaml",
			wantError:   false,
			wantKind:    "GlobalConfig",
			wantGroup:   "config.example.com",
			wantVersion: "v1",
			wantPlural:  "globalconfigs",
			wantScope:   "Cluster",
		},
		{
			name:        "multi-version CRD",
			filename:    "../../test/fixtures/multi-version-crd.yaml",
			wantError:   false,
			wantKind:    "Database",
			wantGroup:   "storage.example.com",
			wantVersion: "v1", // storage version
			wantPlural:  "databases",
			wantScope:   "Namespaced",
		},
		{
			name:      "non-existent file",
			filename:  "nonexistent.yaml",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := analyzer.ParseCRDFromFile(tt.filename)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, info)

			assert.Equal(t, tt.wantKind, info.Kind)
			assert.Equal(t, tt.wantGroup, info.Group)
			assert.Equal(t, tt.wantPlural, info.Plural)
			assert.Equal(t, tt.wantVersion, info.Version)
			assert.Equal(t, tt.wantScope, string(info.CRD.Spec.Scope))
		})
	}
}

func TestCRDInfoMethods(t *testing.T) {
	analyzer := NewCRDAnalyzer()

	info, err := analyzer.ParseCRDFromFile("../../test/fixtures/simple-crd.yaml")
	require.NoError(t, err)
	require.NotNil(t, info)

	// Test package name generation
	packageName := info.GetPackageName()
	assert.Equal(t, "widgets", packageName)

	// Test type name generation
	typeName := info.GetTypeName()
	assert.Equal(t, "Widget", typeName)

	// Test list type name generation
	listTypeName := info.GetListTypeName()
	assert.Equal(t, "WidgetList", listTypeName)

	// Test API version
	apiVersion := info.GetAPIVersion()
	assert.Equal(t, "example.com/v1", apiVersion)

	// Test group-version-kind (uses Kubernetes standard format)
	gvk := info.GetGroupVersionKind()
	assert.Equal(t, "example.com/v1, Kind=Widget", gvk)
}

func TestNewCRDAnalyzer(t *testing.T) {
	analyzer := NewCRDAnalyzer()
	require.NotNil(t, analyzer)
}

func TestParseInvalidFiles(t *testing.T) {
	analyzer := NewCRDAnalyzer()

	tests := []struct {
		name     string
		filename string
	}{
		{"non-existent file", "nonexistent.yaml"},
		{"not a YAML file", "../../pkg/analyzer/crd.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := analyzer.ParseCRDFromFile(tt.filename)
			assert.Error(t, err)
		})
	}
}

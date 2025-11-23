// Command mcp-toolgen generates Go toolsets from Kubernetes Custom Resource Definitions (CRDs)
// for integration with the extendable-kubernetes-mcp-server.
//
// Usage:
//
//	mcp-toolgen --crd <crd-file> --output <output-dir> --package <package-name> --module-path <module>
//	mcp-toolgen --crd-dir <crd-dir> --output-base <base-dir> --module-path <module>
//
// For more information, see https://github.com/friedrichwilken/mcp-toolgen
package main

import (
	"fmt"
	"os"

	"github.com/friedrichwilken/mcp-toolgen/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Package cmd provides the CLI interface for mcp-toolgen using Cobra.
// It handles command-line arguments, flags, and orchestrates the CRD analysis and code generation process.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/friedrichwilken/mcp-toolgen/pkg/analyzer"
	"github.com/friedrichwilken/mcp-toolgen/pkg/generator"
)

var (
	cfgFile             string
	verbose             bool
	dryRun              bool
	overwrite           bool
	crudOperations      string
	crdFile             string
	crdDir              string
	outputDir           string
	outputBase          string
	packageName         string
	modulePath          string
	templateDir         string
	registerToolset     bool
	modulesFilePath     string
	generateCRDResource bool
	generateDocResource string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcp-toolgen",
	Short: "Generate Go toolsets from Kubernetes CRDs",
	Long: `MCP Toolgen is a code generation tool that creates Go toolsets
for Kubernetes Custom Resource Definitions (CRDs). It generates complete
MCP-compatible packages with CRUD operations, client wrappers, and validation schemas.

The tool analyzes CRD YAML files and produces Go code that seamlessly integrates
with the extendable-kubernetes-mcp-server architecture.`,
	Example: `  # Generate toolset from a single CRD
  mcp-toolgen --crd ./crds/function-crd.yaml --output ./pkg/functions --package functions

  # Generate toolsets from a directory of CRDs
  mcp-toolgen --crd-dir ./crds --output-base ./pkg

  # Generate with custom module path
  mcp-toolgen --crd ./crds/function-crd.yaml --output ./pkg/functions --module-path github.com/myorg/myproject

  # Generate only create and read operations
  mcp-toolgen --crud cr --crd ./crds/function-crd.yaml --output ./pkg/functions

  # Generate only delete operations
  mcp-toolgen --crud d --crd ./crds/function-crd.yaml --output ./pkg/functions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mcp-toolgen.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print what would be generated without creating files")

	// Input flags
	rootCmd.Flags().StringVar(&crdFile, "crd", "", "path to CRD YAML file")
	rootCmd.Flags().StringVar(&crdDir, "crd-dir", "", "directory containing CRD YAML files")

	// Output flags
	rootCmd.Flags().StringVar(&outputDir, "output", "", "output directory for generated code")
	rootCmd.Flags().StringVar(&outputBase, "output-base", "", "base directory for multi-CRD generation (creates subdirectories)")
	rootCmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing files")

	// Generation flags
	rootCmd.Flags().StringVar(&packageName, "package", "", "Go package name (defaults to CRD plural name)")
	rootCmd.Flags().StringVar(&modulePath, "module-path", "github.com/example/project", "Go module path")
	rootCmd.Flags().StringVar(&templateDir, "templates", "", "custom template directory (optional)")
	rootCmd.Flags().StringVar(&crudOperations, "crud", "crud", "CRUD operations to generate (c=create, r=read, u=update, d=delete)")
	rootCmd.Flags().BoolVar(&generateCRDResource, "generate-crd-resource", false,
		"generate MCP resource for CRD definition (requires ek8sms with resource support)")
	rootCmd.Flags().StringVar(&generateDocResource, "generate-doc-resource", "",
		"generate MCP resource for documentation (file path or URL, e.g., ./docs.md or https://raw.githubusercontent.com/...)")

	// Registration flags
	rootCmd.Flags().BoolVar(&registerToolset, "register", false, "automatically add import to modules.go after generation")
	rootCmd.Flags().StringVar(&modulesFilePath, "modules-file", "", "path to modules.go file (defaults to <target-repo>/pkg/mcp/modules.go)")

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("module-path") // Error only if flag doesn't exist (programming error)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mcp-toolgen")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// runGenerate executes the main generation logic
func runGenerate() error {
	// Validate input flags
	if err := validateFlags(); err != nil {
		return err
	}

	if crdFile != "" {
		// Generate from single CRD
		return generateFromSingleCRD()
	} else if crdDir != "" {
		// Generate from directory of CRDs
		return generateFromDirectory()
	}

	return fmt.Errorf("either --crd or --crd-dir must be specified")
}

// validateFlags validates the command line flags
func validateFlags() error {
	if crdFile == "" && crdDir == "" {
		return fmt.Errorf("either --crd or --crd-dir must be specified")
	}

	if crdFile != "" && crdDir != "" {
		return fmt.Errorf("--crd and --crd-dir are mutually exclusive")
	}

	if crdFile != "" && outputDir == "" {
		return fmt.Errorf("--output is required when using --crd")
	}

	if crdDir != "" && outputBase == "" {
		return fmt.Errorf("--output-base is required when using --crd-dir")
	}

	if modulePath == "" {
		return fmt.Errorf("--module-path is required")
	}

	// Validate CRUD operations
	if err := validateCRUDOperations(crudOperations); err != nil {
		return fmt.Errorf("invalid --crud flag: %w", err)
	}

	return nil
}

// validateCRUDOperations validates the CRUD operations string
func validateCRUDOperations(crud string) error {
	if crud == "" {
		return fmt.Errorf("CRUD operations cannot be empty")
	}

	validChars := map[rune]bool{
		'c': true, // create
		'r': true, // read (get + list)
		'u': true, // update
		'd': true, // delete
	}

	seen := make(map[rune]bool)
	for _, char := range crud {
		if !validChars[char] {
			return fmt.Errorf("invalid character '%c', valid characters are: c, r, u, d", char)
		}
		if seen[char] {
			return fmt.Errorf("duplicate character '%c' in CRUD operations", char)
		}
		seen[char] = true
	}

	return nil
}

// parseCRUDOperations converts CRUD string to operation slice
func parseCRUDOperations(crud string) []string {
	var operations []string

	for _, char := range crud {
		switch char {
		case 'c':
			operations = append(operations, "create")
		case 'r':
			// Read includes both get and list operations
			operations = append(operations, "get", "list")
		case 'u':
			operations = append(operations, "update")
		case 'd':
			operations = append(operations, "delete")
		}
	}

	// Remove duplicates (in case user specifies 'r' which adds both get and list)
	seen := make(map[string]bool)
	var uniqueOps []string
	for _, op := range operations {
		if !seen[op] {
			uniqueOps = append(uniqueOps, op)
			seen[op] = true
		}
	}

	return uniqueOps
}

// generateFromSingleCRD generates code from a single CRD file
func generateFromSingleCRD() error {
	if verbose {
		fmt.Printf("Generating toolset from CRD: %s\n", crdFile)
		fmt.Printf("Output directory: %s\n", outputDir)
	}

	// Parse CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	crdInfo, err := crdAnalyzer.ParseCRDFromFile(crdFile)
	if err != nil {
		return fmt.Errorf("failed to parse CRD file %s: %w", crdFile, err)
	}

	if verbose {
		fmt.Printf("Parsed CRD: %s (%s)\n", crdInfo.Kind, crdInfo.GetAPIVersion())
	}

	// Load documentation if requested
	if generateDocResource != "" {
		if verbose {
			fmt.Printf("Loading documentation from: %s\n", generateDocResource)
		}
		docContent, err := analyzer.LoadDocumentationContent(generateDocResource)
		if err != nil {
			return fmt.Errorf("failed to load documentation: %w", err)
		}
		crdInfo.DocContent = docContent
		if verbose {
			fmt.Printf("Loaded documentation: %d bytes\n", len(docContent))
		}
	}

	// Create generation config
	config := analyzer.DefaultGenerationConfig()
	config.PackageName = packageName
	config.ModulePath = modulePath
	config.OutputDir = outputDir
	config.TemplateDir = templateDir
	config.SelectedOperations = parseCRUDOperations(crudOperations)
	config.GenerateCRDResource = generateCRDResource
	config.GenerateDocResource = generateDocResource != ""
	config.DocResourcePath = generateDocResource

	if config.PackageName == "" {
		config.PackageName = crdInfo.GetPackageName()
	}

	if verbose {
		fmt.Printf("Selected CRUD operations: %v\n", config.SelectedOperations)
	}

	// Create toolset info
	toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
	if err != nil {
		return fmt.Errorf("failed to create toolset info: %w", err)
	}

	// Generate code
	return generateToolset(toolsetInfo, outputDir)
}

// generateFromDirectory generates code from all CRD files in a directory
func generateFromDirectory() error {
	if verbose {
		fmt.Printf("Generating toolsets from directory: %s\n", crdDir)
		fmt.Printf("Output base directory: %s\n", outputBase)
	}

	// Find all CRD files
	crdFiles, err := findCRDFiles(crdDir)
	if err != nil {
		return fmt.Errorf("failed to find CRD files: %w", err)
	}

	if len(crdFiles) == 0 {
		return fmt.Errorf("no CRD files found in directory %s", crdDir)
	}

	if verbose {
		fmt.Printf("Found %d CRD files\n", len(crdFiles))
	}

	// Generate toolset for each CRD
	crdAnalyzer := analyzer.NewCRDAnalyzer()
	for _, crdFile := range crdFiles {
		if verbose {
			fmt.Printf("Processing %s...\n", crdFile)
		}

		// Parse CRD
		crdInfo, err := crdAnalyzer.ParseCRDFromFile(crdFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", crdFile, err)
			continue
		}

		// Load documentation if requested
		if generateDocResource != "" {
			if verbose {
				fmt.Printf("Loading documentation from: %s\n", generateDocResource)
			}
			docContent, err := analyzer.LoadDocumentationContent(generateDocResource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load documentation for %s: %v\n", crdFile, err)
				continue
			}
			crdInfo.DocContent = docContent
			if verbose {
				fmt.Printf("Loaded documentation: %d bytes\n", len(docContent))
			}
		}

		// Create output directory for this CRD
		packageName := crdInfo.GetPackageName()
		crdOutputDir := filepath.Join(outputBase, packageName)

		// Create generation config
		config := analyzer.DefaultGenerationConfig()
		config.PackageName = packageName
		config.ModulePath = modulePath
		config.OutputDir = crdOutputDir
		config.TemplateDir = templateDir
		config.SelectedOperations = parseCRUDOperations(crudOperations)
		config.GenerateCRDResource = generateCRDResource
		config.GenerateDocResource = generateDocResource != ""
		config.DocResourcePath = generateDocResource

		// Create toolset info
		toolsetInfo, err := analyzer.NewToolsetInfo(crdInfo, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create toolset info for %s: %v\n", crdFile, err)
			continue
		}

		// Generate code
		if err := generateToolset(toolsetInfo, crdOutputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate toolset for %s: %v\n", crdFile, err)
			continue
		}

		if verbose {
			fmt.Printf("Generated toolset for %s in %s\n", crdInfo.Kind, crdOutputDir)
		}
	}

	return nil
}

// generateToolset generates a complete toolset
func generateToolset(toolsetInfo *analyzer.ToolsetInfo, outputDir string) error {
	// Create generator config
	genConfig := &generator.GeneratorConfig{
		OutputDir:       outputDir,
		TemplateDir:     templateDir,
		PackageName:     toolsetInfo.PackageName,
		ModulePath:      modulePath,
		OverwriteFiles:  overwrite,
		IncludeComments: true,
	}

	// Create generator
	gen, err := generator.NewGenerator(genConfig)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Dry run check
	if dryRun {
		fmt.Printf("Would generate toolset for %s in %s\n", toolsetInfo.CRD.Kind, outputDir)
		fmt.Printf("Package: %s\n", toolsetInfo.PackageName)
		fmt.Printf("Files: toolset.go, types.go, client.go, handlers.go, schema.go, doc.go\n")
		return nil
	}

	// Generate toolset
	if err := gen.GenerateToolset(toolsetInfo); err != nil {
		return fmt.Errorf("failed to generate toolset: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully generated toolset in %s\n", outputDir)
	}

	// Register toolset if --register flag is set
	if registerToolset {
		if err := registerToolsetImport(toolsetInfo.PackageName, outputDir); err != nil {
			return fmt.Errorf("failed to register toolset: %w", err)
		}
		if verbose {
			fmt.Printf("Successfully registered toolset in modules.go\n")
		}
	}

	return nil
}

// findCRDFiles finds all YAML files in a directory that could be CRDs
func findCRDFiles(dir string) ([]string, error) {
	var crdFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check for YAML files
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			crdFiles = append(crdFiles, path)
		}

		return nil
	})

	return crdFiles, err
}

// registerToolsetImport adds the generated toolset import to modules.go
func registerToolsetImport(packageName, outputDir string) error {
	// Determine modules.go location
	modulesPath, err := generator.DetermineModulesFilePath(outputDir, modulePath, modulesFilePath)
	if err != nil {
		return err
	}

	// Construct import path
	importPath := filepath.Join(modulePath, "pkg", packageName)

	if verbose {
		fmt.Printf("Registering toolset: %s\n", importPath)
		fmt.Printf("In modules file: %s\n", modulesPath)
	}

	// Register the import
	return generator.RegisterInModulesFile(modulesPath, importPath)
}

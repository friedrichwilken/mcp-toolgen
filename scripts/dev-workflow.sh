#!/bin/bash

# Reverse-Kube Builder Development Workflow Script
# This script provides common development workflows and automation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Show help
show_help() {
    cat << EOF
Reverse-Kube Builder Development Workflow

Usage: $0 [COMMAND]

Commands:
    setup       Set up development environment
    dev         Quick development cycle (build + basic test)
    full-test   Run complete test suite with all demos
    demo        Run all demo scenarios
    benchmark   Run performance benchmarks
    clean       Clean all generated files
    release     Prepare for release (QA + multi-platform build)
    integrate   Generate toolset for ek8sms integration
    help        Show this help message

Examples:
    $0 setup      # Set up development environment
    $0 dev        # Quick development cycle
    $0 demo       # Run all demos
    $0 integrate  # Generate for ek8sms integration

EOF
}

# Set up development environment
setup_dev() {
    log_info "Setting up development environment..."
    make dev-setup
    log_success "Development environment ready!"
}

# Quick development cycle
quick_dev() {
    log_info "Running quick development cycle..."

    log_info "Building binary..."
    make dev-build

    log_info "Running basic tests..."
    make dev-test

    log_info "Validating CRUD functionality..."
    make validate-crud

    log_success "Quick development cycle completed!"
}

# Full test suite
full_test() {
    log_info "Running complete test suite..."

    log_info "Quality assurance checks..."
    make qa

    log_info "Building binary..."
    make build

    log_info "Running all demos..."
    make docs-examples

    log_info "Performance benchmarks..."
    make benchmark

    log_success "Full test suite completed!"
}

# Run all demos
run_demos() {
    log_info "Running all demo scenarios..."

    log_info "Demo 1: Create + Read operations..."
    make demo-cr

    log_info "Demo 2: Delete operations only..."
    make demo-d

    log_info "Demo 3: All CRUD operations..."
    make demo-all

    log_info "Demo 4: Dry run..."
    make demo-dry-run

    log_info "Checking generated examples..."
    echo "Generated examples:"
    ls -la examples/ 2>/dev/null || log_warning "No examples directory found"

    log_success "All demos completed!"
}

# Run benchmarks
run_benchmark() {
    log_info "Running performance benchmarks..."

    echo "🚀 Starting benchmark run..."
    make benchmark

    log_info "Benchmark details:"
    echo "  - CRD: Kyma Serverless Functions (34KB)"
    echo "  - Operations: Full CRUD generation"
    echo "  - Output: Complete Go toolset (6 files)"

    log_success "Benchmark completed!"
}

# Clean all generated files
clean_all() {
    log_info "Cleaning all generated files..."

    make clean
    make clean-examples

    # Additional cleanup
    rm -rf /tmp/benchmark /tmp/test

    log_success "Cleanup completed!"
}

# Prepare for release
prepare_release() {
    log_info "Preparing for release..."

    log_info "Running CI pipeline..."
    make ci

    log_info "Building for all platforms..."
    make build-all

    log_info "Checking build artifacts..."
    ls -la build/

    log_success "Release preparation completed!"
}

# Generate for ek8sms integration
integrate_ek8sms() {
    log_info "Generating toolset for ek8sms integration..."

    if [ ! -f "/Users/I549741/claude-playroom/crds/kyma-serverless-function-crd.yaml" ]; then
        log_error "Kyma Functions CRD not found!"
        exit 1
    fi

    if [ ! -d "/Users/I549741/claude-playroom/extendable-kubernetes-mcp-server" ]; then
        log_error "ek8sms directory not found!"
        exit 1
    fi

    make generate-example

    log_info "Generated files in ek8sms:"
    ls -la /Users/I549741/claude-playroom/extendable-kubernetes-mcp-server/pkg/functions/ 2>/dev/null || log_warning "Functions package not found"

    log_success "ek8sms integration ready!"
}

# Main command handling
case "${1:-help}" in
    setup)
        setup_dev
        ;;
    dev)
        quick_dev
        ;;
    full-test)
        full_test
        ;;
    demo)
        run_demos
        ;;
    benchmark)
        run_benchmark
        ;;
    clean)
        clean_all
        ;;
    release)
        prepare_release
        ;;
    integrate)
        integrate_ek8sms
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        echo
        show_help
        exit 1
        ;;
esac
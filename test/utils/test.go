// Package utils provides common testing utilities for mcp-toolgen.
package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Must is a helper function that panics if an error is not nil.
// Useful for test setup where failures should immediately fail the test.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// ReadFile reads a file relative to the caller's location and returns content as string.
// This is useful for loading test fixtures and test data files.
func ReadFile(path ...string) string {
	_, file, _, _ := runtime.Caller(1)
	filePath := filepath.Join(append([]string{filepath.Dir(file)}, path...)...)
	fileBytes := Must(os.ReadFile(filePath)) // #nosec G304 -- test helper for reading test fixtures
	return string(fileBytes)
}

// ReadFileBytes reads a file and returns content as bytes.
func ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path) // #nosec G304 -- test helper for reading test fixtures
}

// RandomPortAddress finds a random available TCP port.
// Returns a TCPAddr that can be used for test servers.
func RandomPortAddress() (*net.TCPAddr, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to find random port: %w", err)
	}
	defer func() { _ = ln.Close() }()
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to cast listener address to TCPAddr")
	}
	return tcpAddr, nil
}

// WaitForServer waits for a server to become available at the given address.
// Useful for integration tests that need to wait for server startup.
func WaitForServer(tcpAddr *net.TCPAddr, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for server at %s", tcpAddr.String())
}

// SkipIfShort skips the test if running in short mode.
// Use this for longer-running integration and e2e tests.
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
}

// TempDir creates a temporary directory for the test.
// The directory will be automatically cleaned up when the test completes.
func TempDir(t *testing.T) string {
	dir := Must(os.MkdirTemp("", "mcp-toolgen-test-"))
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

// WriteTestFile writes content to a file in the given directory.
// Returns the full path to the created file.
func WriteTestFile(t *testing.T, dir, filename, content string) string {
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file %s: %v", path, err)
	}
	return path
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists at the given path.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

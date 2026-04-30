package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetModuleName(t *testing.T) {
	name := getModuleName()

	if name != "github.com/FFengIll/sortimport" {
		t.Errorf("expected github.com/FFengIll/sortimport, got: %s", name)
	}
}

func TestFindModulePath(t *testing.T) {
	// Test finding module from current directory
	modulePath := findModulePath(".")
	if modulePath != "github.com/FFengIll/sortimport" {
		t.Errorf("expected github.com/FFengIll/sortimport, got: %s", modulePath)
	}
}

func TestFindModulePath_FromFile(t *testing.T) {
	// Test finding module from a file path
	modulePath := findModulePath("sortimport_test.go")
	if modulePath != "github.com/FFengIll/sortimport" {
		t.Errorf("expected github.com/FFengIll/sortimport, got: %s", modulePath)
	}
}

func TestFindModulePath_NonExistent(t *testing.T) {
	// Test from a path that doesn't exist (should still work by traversing up)
	// Using a non-existent nested path
	modulePath := findModulePath("/tmp/nonexistent/deep/path")
	// May or may not find a go.mod, but should not crash
	_ = modulePath
}

func TestFindModulePath_NestedDir(t *testing.T) {
	// Create a temp directory structure to test nested directory detection
	tmpDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Create go.mod in the root of tmpDir
	goModContent := `module example.com/testmodule

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Test finding module from nested directory
	modulePath := findModulePath(nestedDir)
	if modulePath != "example.com/testmodule" {
		t.Errorf("expected example.com/testmodule, got: %s", modulePath)
	}

	// Test finding module from a file in nested directory
	testFile := filepath.Join(nestedDir, "test.go")
	modulePath = findModulePath(testFile)
	if modulePath != "example.com/testmodule" {
		t.Errorf("expected example.com/testmodule, got: %s", modulePath)
	}
}

func TestIsLocalPackageWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		impName  string
		prefix   string
		expected bool
	}{
		{
			name:     "match with quotes",
			impName:  `"github.com/user/project/pkg"`,
			prefix:   "github.com/user/project",
			expected: true,
		},
		{
			name:     "match without quotes",
			impName:  "github.com/user/project/pkg",
			prefix:   "github.com/user/project",
			expected: true,
		},
		{
			name:     "no match different module",
			impName:  `"github.com/other/project"`,
			prefix:   "github.com/user/project",
			expected: false,
		},
		{
			name:     "empty prefix",
			impName:  `"github.com/user/project"`,
			prefix:   "",
			expected: false,
		},
		{
			name:     "exact match",
			impName:  `"github.com/user/project"`,
			prefix:   "github.com/user/project",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalPackageWithPrefix(tt.impName, tt.prefix)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

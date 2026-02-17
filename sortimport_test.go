package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Load standard packages before running tests
	if err := loadStandardPackages(); err != nil {
		panic("failed to load standard packages: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestProcessFile(t *testing.T) {
	*localPrefix = "github.com/AanZee/goimportssort"
	reader := strings.NewReader(`package main

// builtin
// external
// local
import (
	"fmt"
	"log"

	APA "bitbucket.org/example/package/name"
	APZ "bitbucket.org/example/package/name"
	"bitbucket.org/example/package/name2"
	"bitbucket.org/example/package/name3" // foopsie
	"bitbucket.org/example/package/name4"

	"github.com/AanZee/goimportssort/package1"
	// a
	"github.com/AanZee/goimportssort/package2"

	/*
		mijn comment
	*/
	"net/http/httptest"
	"database/sql/driver"
)
// klaslkasdko

func main() {
	fmt.Println("Hello!")
}`)
	want := `package main

import (
	"database/sql/driver"
	"fmt"
	"log"
	"net/http/httptest"

	APA "bitbucket.org/example/package/name"
	APZ "bitbucket.org/example/package/name"
	"bitbucket.org/example/package/name2"
	"bitbucket.org/example/package/name3"
	"bitbucket.org/example/package/name4"

	"github.com/AanZee/goimportssort/package1"
	"github.com/AanZee/goimportssort/package2"
)

func main() {
	fmt.Println("Hello!")
}
`

	output, err := processFile("", reader, os.Stdout)
	if output == nil {
		t.Error("expected non-nil output")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestProcessFile_SingleImport(t *testing.T) {
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(
		`package main


import "github.com/AanZee/goimportssort/package1"


func main() {
	fmt.Println("Hello!")
}`)
	want := `package main

import (
	"github.com/AanZee/goimportssort/package1"
)

func main() {
	fmt.Println("Hello!")
}
`
	output, err := processFile("", reader, os.Stdout)
	if output == nil {
		t.Error("expected non-nil output")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestProcessFile_EmptyImport(t *testing.T) {
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(`package main

func main() {
	fmt.Println("Hello!")
}`)
	want := `package main

func main() {
	fmt.Println("Hello!")
}`
	output, err := processFile("", reader, os.Stdout)
	if output == nil {
		t.Error("expected non-nil output")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestProcessFile_ReadMeExample(t *testing.T) {
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(`package main

import (
	"fmt"
	"log"
	APZ "bitbucket.org/example/package/name"
	APA "bitbucket.org/example/package/name"
	"github.com/AanZee/goimportssort/package2"
	"github.com/AanZee/goimportssort/package1"
)
import (
	"net/http/httptest"
)

import "bitbucket.org/example/package/name2"
import "bitbucket.org/example/package/name3"
import "bitbucket.org/example/package/name4"`)
	want := `package main

import (
	"fmt"
	"log"
	"net/http/httptest"

	APA "bitbucket.org/example/package/name"
	APZ "bitbucket.org/example/package/name"
	"bitbucket.org/example/package/name2"
	"bitbucket.org/example/package/name3"
	"bitbucket.org/example/package/name4"

	"github.com/AanZee/goimportssort/package1"
	"github.com/AanZee/goimportssort/package2"
)
`
	output, err := processFile("", reader, os.Stdout)
	if output == nil {
		t.Error("expected non-nil output")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestProcessFile_WronglyFormattedGo(t *testing.T) {
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(
		`package main
import "github.com/AanZee/goimportssort/package1"


func main() {
	fmt.Println("Hello!")
}`)
	want := `package main

import (
	"github.com/AanZee/goimportssort/package1"
)

func main() {
	fmt.Println("Hello!")
}
`
	output, err := processFile("", reader, os.Stdout)
	if output == nil {
		t.Error("expected non-nil output")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestGetModuleName(t *testing.T) {
	name := getModuleName()

	if name != "github.com/FFengIll/sortimport" {
		t.Errorf("expected github.com/FFengIll/sortimport, got: %s", name)
	}
}

func Test_loadStandardPackages(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "load",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadStandardPackages(); (err != nil) != tt.wantErr {
				t.Errorf("loadStandardPackages() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCacheManager_New(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if cm == nil {
		t.Error("expected non-nil cache manager")
	}
	if cm.version == "" {
		t.Error("expected non-empty version")
	}
	if !strings.Contains(cm.cacheDir, ".cache/sortimport") {
		t.Errorf("expected cacheDir to contain .cache/sortimport, got: %s", cm.cacheDir)
	}
}

func TestCacheManager_GetCacheFile(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	cacheFile := cm.getCacheFile()
	if !strings.Contains(cacheFile, ".cache/sortimport") {
		t.Errorf("expected cacheFile to contain .cache/sortimport, got: %s", cacheFile)
	}
	if !strings.Contains(cacheFile, cm.version) {
		t.Errorf("expected cacheFile to contain version %s, got: %s", cm.version, cacheFile)
	}
	if !strings.HasSuffix(cacheFile, ".json") {
		t.Errorf("expected cacheFile to end with .json, got: %s", cacheFile)
	}
}

func TestCacheManager_WriteAndRead(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	cm := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.21.0",
	}

	// Test data
	testPackages := map[string]struct{}{
		"fmt":     {},
		"os":      {},
		"strings": {},
	}

	// Write cache
	err := cm.write(testPackages)
	if err != nil {
		t.Errorf("expected no error on write, got: %v", err)
	}

	// Verify file exists
	cacheFile := cm.getCacheFile()
	_, err = os.Stat(cacheFile)
	if err != nil {
		t.Errorf("expected cache file to exist, got error: %v", err)
	}

	// Read cache
	info, err := cm.read()
	if err != nil {
		t.Errorf("expected no error on read, got: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil cache info")
	}
	if info.Version != "go1.21.0" {
		t.Errorf("expected version go1.21.0, got: %s", info.Version)
	}
	if _, ok := info.Data["fmt"]; !ok {
		t.Error("expected fmt in cache data")
	}
	if _, ok := info.Data["os"]; !ok {
		t.Error("expected os in cache data")
	}
	if _, ok := info.Data["strings"]; !ok {
		t.Error("expected strings in cache data")
	}
}

func TestCacheManager_ReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.99.0", // Non-existent version
	}

	info, err := cm.read()
	if err == nil {
		t.Error("expected error for non-existent cache")
	}
	if info != nil {
		t.Errorf("expected nil info, got: %v", info)
	}
}

func TestCacheManager_VersionIndependent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache manager for go1.21.0
	cm1 := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.21.0",
	}
	packages1 := map[string]struct{}{"fmt": {}}
	err := cm1.write(packages1)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Create cache manager for go1.22.0
	cm2 := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.22.0",
	}
	packages2 := map[string]struct{}{"os": {}, "io": {}}
	err = cm2.write(packages2)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Both cache files should exist
	if _, err := os.Stat(cm1.getCacheFile()); os.IsNotExist(err) {
		t.Error("expected cm1 cache file to exist")
	}
	if _, err := os.Stat(cm2.getCacheFile()); os.IsNotExist(err) {
		t.Error("expected cm2 cache file to exist")
	}

	// Read both and verify they are independent
	info1, err := cm1.read()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if _, ok := info1.Data["fmt"]; !ok {
		t.Error("expected fmt in info1.Data")
	}
	if _, ok := info1.Data["os"]; ok {
		t.Error("did not expect os in info1.Data")
	}

	info2, err := cm2.read()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if _, ok := info2.Data["os"]; !ok {
		t.Error("expected os in info2.Data")
	}
	if _, ok := info2.Data["io"]; !ok {
		t.Error("expected io in info2.Data")
	}
	if _, ok := info2.Data["fmt"]; ok {
		t.Error("did not expect fmt in info2.Data")
	}
}

func TestCacheManager_GetOldCachePath(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	oldPath := cm.getOldCachePath()
	if !strings.Contains(oldPath, ".cache/sortimport.json") {
		t.Errorf("expected oldPath to contain .cache/sortimport.json, got: %s", oldPath)
	}
}

func TestCurrentGoVersionCache(t *testing.T) {
	cm, err := newCacheManager()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// The cache file should include the current Go version
	expectedVersion := runtime.Version()
	if cm.version != expectedVersion {
		t.Errorf("expected version %s, got: %s", expectedVersion, cm.version)
	}

	cacheFile := cm.getCacheFile()
	if !strings.Contains(cacheFile, expectedVersion) {
		t.Errorf("expected cacheFile to contain %s, got: %s", expectedVersion, cacheFile)
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

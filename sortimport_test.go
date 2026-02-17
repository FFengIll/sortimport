package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Load standard packages before running tests
	if err := loadStandardPackages(); err != nil {
		panic("failed to load standard packages: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestProcessFile(t *testing.T) {
	asserts := assert.New(t)
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
	asserts.NotEqual(nil, output)
	asserts.Equal(nil, err)
	asserts.Equal(want, string(output))
}

func TestProcessFile_SingleImport(t *testing.T) {
	asserts := assert.New(t)
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(
		`package main


import "github.com/AanZee/goimportssort/package1"


func main() {
	fmt.Println("Hello!")
}`)
	output, err := processFile("", reader, os.Stdout)
	asserts.NotEqual(nil, output)
	asserts.Equal(nil, err)
	asserts.Equal(
		`package main

import (
	"github.com/AanZee/goimportssort/package1"
)

func main() {
	fmt.Println("Hello!")
}
`, string(output))
}

func TestProcessFile_EmptyImport(t *testing.T) {
	asserts := assert.New(t)
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(`package main

func main() {
	fmt.Println("Hello!")
}`)
	output, err := processFile("", reader, os.Stdout)
	asserts.NotEqual(nil, output)
	asserts.Equal(nil, err)
	asserts.Equal(`package main

func main() {
	fmt.Println("Hello!")
}`, string(output))
}

func TestProcessFile_ReadMeExample(t *testing.T) {
	asserts := assert.New(t)
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
	output, err := processFile("", reader, os.Stdout)
	asserts.NotEqual(nil, output)
	asserts.Equal(nil, err)
	asserts.Equal(`package main

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
`, string(output))
}

func TestProcessFile_WronglyFormattedGo(t *testing.T) {
	asserts := assert.New(t)
	*localPrefix = "github.com/AanZee/goimportssort"

	reader := strings.NewReader(
		`package main
import "github.com/AanZee/goimportssort/package1"


func main() {
	fmt.Println("Hello!")
}`)
	output, err := processFile("", reader, os.Stdout)
	asserts.NotEqual(nil, output)
	asserts.Equal(nil, err)
	asserts.Equal(
		`package main

import (
	"github.com/AanZee/goimportssort/package1"
)

func main() {
	fmt.Println("Hello!")
}
`, string(output))
}

func TestGetModuleName(t *testing.T) {
	asserts := assert.New(t)

	name := getModuleName()

	asserts.Equal("github.com/FFengIll/sortimport", name)
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
	assert.NoError(t, err)
	assert.NotNil(t, cm)
	assert.NotEmpty(t, cm.version)
	assert.Contains(t, cm.cacheDir, ".cache/sortimport")
}

func TestCacheManager_GetCacheFile(t *testing.T) {
	cm, err := newCacheManager()
	assert.NoError(t, err)

	cacheFile := cm.getCacheFile()
	assert.Contains(t, cacheFile, ".cache/sortimport")
	assert.Contains(t, cacheFile, cm.version)
	assert.True(t, strings.HasSuffix(cacheFile, ".json"))
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
	assert.NoError(t, err)

	// Verify file exists
	cacheFile := cm.getCacheFile()
	_, err = os.Stat(cacheFile)
	assert.NoError(t, err)

	// Read cache
	info, err := cm.read()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "go1.21.0", info.Version)
	assert.Contains(t, info.Data, "fmt")
	assert.Contains(t, info.Data, "os")
	assert.Contains(t, info.Data, "strings")
}

func TestCacheManager_ReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cm := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.99.0", // Non-existent version
	}

	info, err := cm.read()
	assert.Error(t, err)
	assert.Nil(t, info)
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
	assert.NoError(t, err)

	// Create cache manager for go1.22.0
	cm2 := &CacheManager{
		cacheDir: tmpDir,
		version:  "go1.22.0",
	}
	packages2 := map[string]struct{}{"os": {}, "io": {}}
	err = cm2.write(packages2)
	assert.NoError(t, err)

	// Both cache files should exist
	assert.FileExists(t, cm1.getCacheFile())
	assert.FileExists(t, cm2.getCacheFile())

	// Read both and verify they are independent
	info1, err := cm1.read()
	assert.NoError(t, err)
	assert.Contains(t, info1.Data, "fmt")
	assert.NotContains(t, info1.Data, "os")

	info2, err := cm2.read()
	assert.NoError(t, err)
	assert.Contains(t, info2.Data, "os")
	assert.Contains(t, info2.Data, "io")
	assert.NotContains(t, info2.Data, "fmt")
}

func TestCacheManager_GetOldCachePath(t *testing.T) {
	cm, err := newCacheManager()
	assert.NoError(t, err)

	oldPath := cm.getOldCachePath()
	assert.Contains(t, oldPath, ".cache/sortimport.json")
}

func TestCurrentGoVersionCache(t *testing.T) {
	cm, err := newCacheManager()
	assert.NoError(t, err)

	// The cache file should include the current Go version
	expectedVersion := runtime.Version()
	assert.Equal(t, expectedVersion, cm.version)

	cacheFile := cm.getCacheFile()
	assert.Contains(t, cacheFile, expectedVersion)
}

func TestFindModulePath(t *testing.T) {
	// Test finding module from current directory
	modulePath := findModulePath(".")
	assert.Equal(t, "github.com/FFengIll/sortimport", modulePath)
}

func TestFindModulePath_FromFile(t *testing.T) {
	// Test finding module from a file path
	modulePath := findModulePath("sortimport_test.go")
	assert.Equal(t, "github.com/FFengIll/sortimport", modulePath)
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
	assert.NoError(t, err)

	// Create go.mod in the root of tmpDir
	goModContent := `module example.com/testmodule

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	assert.NoError(t, err)

	// Test finding module from nested directory
	modulePath := findModulePath(nestedDir)
	assert.Equal(t, "example.com/testmodule", modulePath)

	// Test finding module from a file in nested directory
	testFile := filepath.Join(nestedDir, "test.go")
	modulePath = findModulePath(testFile)
	assert.Equal(t, "example.com/testmodule", modulePath)
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
			assert.Equal(t, tt.expected, result)
		})
	}
}

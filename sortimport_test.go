package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dave/dst/decorator"
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

// resetFlag restores the value pointed to by *string flag pointers after a test.
func resetStringFlag(t *testing.T, ptr *string) {
	t.Helper()
	prev := *ptr
	t.Cleanup(func() { *ptr = prev })
}

func resetBoolFlag(t *testing.T, ptr *bool) {
	t.Helper()
	prev := *ptr
	t.Cleanup(func() { *ptr = prev })
}

func TestProcessFile_SecondPart(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetStringFlag(t, secondPrefix)
	*localPrefix = "github.com/myorg/myrepo"
	*secondPrefix = "github.com/myorg"

	reader := strings.NewReader(`package main

import (
	"fmt"
	"github.com/external/lib"
	"github.com/myorg/shared"
	"github.com/myorg/myrepo/pkg"
	"os"
)

func main() {}
`)
	want := `package main

import (
	"fmt"
	"os"

	"github.com/external/lib"

	"github.com/myorg/shared"

	"github.com/myorg/myrepo/pkg"
)

func main() {}
`

	output, err := processFile("", reader, os.Stdout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(output) != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, string(output))
	}
}

func TestProcessFile_BlankAndDotImport(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetStringFlag(t, secondPrefix)
	*localPrefix = "github.com/myorg/myrepo"

	reader := strings.NewReader(`package main

import (
	_ "embed"
	. "fmt"
	mylog "log"
	"os"
)

func main() {}
`)

	output, err := processFile("", reader, os.Stdout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	got := string(output)
	for _, snippet := range []string{
		`_ "embed"`,
		`. "fmt"`,
		`mylog "log"`,
		`"os"`,
	} {
		if !strings.Contains(got, snippet) {
			t.Errorf("expected output to contain %q, got:\n%s", snippet, got)
		}
	}
}

func TestProcessFile_InvalidSource(t *testing.T) {
	resetStringFlag(t, localPrefix)
	*localPrefix = "github.com/myorg/myrepo"

	// Missing closing brace makes this unparseable.
	reader := strings.NewReader(`package main
import "fmt"
func main() { fmt.Println("Hello"
`)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("processFile should not panic on invalid source, got: %v", r)
		}
	}()

	_, err := processFile("", reader, os.Stdout)
	if err == nil {
		t.Fatal("expected error on invalid source, got nil")
	}
}

func TestProcessFile_ListMode(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, list)
	*localPrefix = "github.com/myorg/myrepo"
	*list = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	dir := t.TempDir()
	fp := filepath.Join(dir, "in.go")
	if err := os.WriteFile(fp, []byte(src), 0644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	var out bytes.Buffer
	res, err := processFile(fp, nil, &out)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(out.String(), "\"fmt\"") || !strings.Contains(out.String(), "\"os\"") {
		t.Errorf("expected list output to contain imports, got: %q", out.String())
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
	// File on disk must NOT be modified in list mode.
	disk, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("read tmp file: %v", err)
	}
	if string(disk) != src {
		t.Errorf("file on disk should not change in list mode\nexpected:\n%s\ngot:\n%s", src, string(disk))
	}
}

func TestProcessFile_WriteMode(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, write)
	*localPrefix = "github.com/myorg/myrepo"
	*write = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	dir := t.TempDir()
	fp := filepath.Join(dir, "in.go")
	if err := os.WriteFile(fp, []byte(src), 0644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	if _, err := processFile(fp, nil, os.Stdout); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	disk, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("read tmp file: %v", err)
	}
	got := string(disk)
	// "fmt" should now precede "os" alphabetically inside the import block.
	fmtIdx := strings.Index(got, "\"fmt\"")
	osIdx := strings.Index(got, "\"os\"")
	if fmtIdx < 0 || osIdx < 0 || fmtIdx > osIdx {
		t.Errorf("write mode should sort imports on disk, got:\n%s", got)
	}
}

func TestProcessFile_PreservesFileMode(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, write)
	*localPrefix = "github.com/myorg/myrepo"
	*write = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	dir := t.TempDir()
	fp := filepath.Join(dir, "in.go")
	if err := os.WriteFile(fp, []byte(src), 0600); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	// Re-chmod to be sure (umask might have stripped perms on creation).
	if err := os.Chmod(fp, 0600); err != nil {
		t.Fatalf("chmod tmp file: %v", err)
	}

	if _, err := processFile(fp, nil, os.Stdout); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	info, err := os.Stat(fp)
	if err != nil {
		t.Fatalf("stat tmp file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("expected mode 0600 preserved, got %o", got)
	}
}

func TestWalkDir(t *testing.T) {
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, list)
	*localPrefix = "github.com/myorg/myrepo"
	// Avoid stdout noise but still exercise the processing branch.
	*list = false

	root := t.TempDir()
	files := map[string]string{
		"a.go":          "package a\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\nvar _ = os.Args\n",
		"b.txt":         "not a go file",
		".hidden.go":    "package hidden\n",
		"sub/c.go":      "package c\nimport (\n\t\"os\"\n\t\"fmt\"\n)\n\nvar _ = fmt.Sprintf\nvar _ = os.Args\n",
		"sub/d.notgo":   "skip me",
		"sub/.dotted.go": "package dotted\n",
	}
	for rel, content := range files {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	if err := walkDir(root); err != nil {
		t.Fatalf("walkDir error: %v", err)
	}
	// We can't easily assert which files were processed without instrumenting,
	// but reaching this point with no error proves the walker handled the mix.
}

func TestProcessPaths_MultiplePaths(t *testing.T) {
	// Verifies the multi-path fix: every path in the slice gets processed,
	// not just the first one.
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, write)
	*localPrefix = "github.com/myorg/myrepo"
	*write = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	dir := t.TempDir()
	paths := []string{
		filepath.Join(dir, "a.go"),
		filepath.Join(dir, "b.go"),
		filepath.Join(dir, "c.go"),
	}
	for _, p := range paths {
		if err := os.WriteFile(p, []byte(src), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	if err := processPaths(paths, os.Stdout); err != nil {
		t.Fatalf("processPaths: %v", err)
	}

	// Every file must be sorted (fmt before os in the import block).
	for _, p := range paths {
		disk, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		fmtIdx := strings.Index(string(disk), "\"fmt\"")
		osIdx := strings.Index(string(disk), "\"os\"")
		if fmtIdx < 0 || osIdx < 0 || fmtIdx > osIdx {
			t.Errorf("file %s not sorted, content:\n%s", p, string(disk))
		}
	}
}

func TestProcessPaths_DirAndFile(t *testing.T) {
	// Mix a directory and a file; both must be processed.
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, write)
	*localPrefix = "github.com/myorg/myrepo"
	*write = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	root := t.TempDir()
	subDir := filepath.Join(root, "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dirFile := filepath.Join(subDir, "in.go")
	loneFile := filepath.Join(root, "lone.go")
	for _, p := range []string{dirFile, loneFile} {
		if err := os.WriteFile(p, []byte(src), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	if err := processPaths([]string{subDir, loneFile}, os.Stdout); err != nil {
		t.Fatalf("processPaths: %v", err)
	}

	for _, p := range []string{dirFile, loneFile} {
		disk, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if strings.Index(string(disk), "\"fmt\"") > strings.Index(string(disk), "\"os\"") {
			t.Errorf("file %s not sorted, content:\n%s", p, string(disk))
		}
	}
}

func TestProcessPaths_StatErrorContinues(t *testing.T) {
	// A nonexistent path must not stop processing of subsequent valid paths.
	resetStringFlag(t, localPrefix)
	resetBoolFlag(t, write)
	*localPrefix = "github.com/myorg/myrepo"
	*write = true

	src := `package main

import (
	"os"
	"fmt"
)

func main() {}
`
	dir := t.TempDir()
	good := filepath.Join(dir, "good.go")
	if err := os.WriteFile(good, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	bad := filepath.Join(dir, "doesnotexist.go")

	err := processPaths([]string{bad, good}, os.Stdout)
	if err == nil {
		t.Error("expected non-nil error for missing path")
	}

	// good.go must still have been processed despite the bad path.
	disk, _ := os.ReadFile(good)
	if strings.Index(string(disk), "\"fmt\"") > strings.Index(string(disk), "\"os\"") {
		t.Errorf("good file not sorted after stat error on earlier path:\n%s", string(disk))
	}
}

func TestConvertImportsToSlice(t *testing.T) {
	resetStringFlag(t, secondPrefix)
	*secondPrefix = "github.com/myorg"

	src := `package main

import (
	"fmt"
	_ "embed"
	"github.com/external/lib"
	"github.com/myorg/shared"
	"github.com/myorg/myrepo/internal/foo"
	alias "github.com/myorg/myrepo/internal/bar"
)
`
	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	mgr, err := convertImportsToSlice(node, "github.com/myorg/myrepo")
	if err != nil {
		t.Fatalf("convertImportsToSlice: %v", err)
	}

	if got := mgr.Standard().countImports(); got != 2 {
		t.Errorf("standard count = %d, want 2", got)
	}
	if got := mgr.ThirdPart().countImports(); got != 1 {
		t.Errorf("third count = %d, want 1", got)
	}
	if got := mgr.SecondPart().countImports(); got != 1 {
		t.Errorf("second count = %d, want 1", got)
	}
	if got := mgr.Local().countImports(); got != 2 {
		t.Errorf("local count = %d, want 2", got)
	}
	if got := mgr.countImports(); got != 6 {
		t.Errorf("total = %d, want 6", got)
	}

	// Ensure the aliased local import preserved its qualifier.
	var foundAlias bool
	for _, m := range mgr.Local().models {
		if m.localReference == "alias" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Error("expected aliased local import to keep its qualifier")
	}
}

func TestIsStandardPackage(t *testing.T) {
	if !isStandardPackage("fmt") {
		t.Error("fmt should be standard")
	}
	if !isStandardPackage("os") {
		t.Error("os should be standard")
	}
	if isStandardPackage("github.com/external/lib") {
		t.Error("third-party package should not be standard")
	}
	if isStandardPackage("") {
		t.Error("empty string should not be standard")
	}
}

func TestCacheManager_Update(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: runtime.Version()}

	if err := cm.update(); err != nil {
		t.Fatalf("update: %v", err)
	}

	cacheFile := cm.getCacheFile()
	if _, err := os.Stat(cacheFile); err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}

	info, err := cm.read()
	if err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if info.Version != runtime.Version() {
		t.Errorf("version = %q, want %q", info.Version, runtime.Version())
	}
	if _, ok := info.Data["fmt"]; !ok {
		t.Error("expected fmt in updated cache data")
	}
}

func TestCacheManager_LoadOrFetch_Hit(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: "go-test-fake"}

	seed := map[string]struct{}{"my/fake/pkg": {}}
	if err := cm.write(seed); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	got, err := cm.loadOrFetch()
	if err != nil {
		t.Fatalf("loadOrFetch: %v", err)
	}
	if _, ok := got["my/fake/pkg"]; !ok {
		t.Error("expected seeded data to be returned (cache hit)")
	}
	// Real std package must NOT be in the seeded result.
	if _, ok := got["fmt"]; ok {
		t.Error("did not expect fmt in seeded cache hit result")
	}
}

func TestCacheManager_LoadOrFetch_Miss(t *testing.T) {
	tmp := t.TempDir()
	cm := &CacheManager{cacheDir: tmp, version: runtime.Version()}

	got, err := cm.loadOrFetch()
	if err != nil {
		t.Fatalf("loadOrFetch: %v", err)
	}
	if _, ok := got["fmt"]; !ok {
		t.Error("expected fetched std packages to include fmt")
	}
	// After miss, cache file should now exist for next call.
	if _, err := os.Stat(cm.getCacheFile()); err != nil {
		t.Errorf("expected cache file after miss: %v", err)
	}
}

func TestSortImports_Stable(t *testing.T) {
	g := &impGroup{models: []*impModel{
		{path: `"github.com/b/x"`},
		{path: `"github.com/a/y"`},
		{path: `"github.com/a/y"`, localReference: "alias"},
		{path: `"github.com/a/x"`},
	}}
	g.sortImports()

	wantPaths := []string{
		`"github.com/a/x"`,
		`"github.com/a/y"`,
		`"github.com/a/y"`,
		`"github.com/b/x"`,
	}
	wantRefs := []string{"", "", "alias", ""}
	for i, m := range g.models {
		if m.path != wantPaths[i] || m.localReference != wantRefs[i] {
			t.Errorf("models[%d] = (%q, %q), want (%q, %q)",
				i, m.path, m.localReference, wantPaths[i], wantRefs[i])
		}
	}
}

func TestConvertImportsToGo_GroupSeparator(t *testing.T) {
	mgr := newImpManager()
	mgr.Standard().append(&impModel{path: `"fmt"`})
	mgr.ThirdPart().append(&impModel{path: `"github.com/x/y"`})
	mgr.Local().append(&impModel{path: `"github.com/myorg/myrepo/pkg"`})

	out := string(mgr.convertImportsToGo())
	if !strings.HasPrefix(out, "import (") {
		t.Errorf("expected import block, got: %q", out)
	}
	if !strings.HasSuffix(out, ")") {
		t.Errorf("expected closing paren, got: %q", out)
	}
	// Group separator: a blank line between non-empty groups.
	// We expect 2 separators (between std/third and third/local).
	if got := strings.Count(out, "\n\n\t"); got != 2 {
		t.Errorf("expected 2 blank-line group separators, got %d in:\n%s", got, out)
	}
}

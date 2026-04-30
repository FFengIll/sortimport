package main

import (
	"os"
	"path/filepath"
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

func TestStripGoEllipsis(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"./...", "."},
		{"./pkg/...", "./pkg"},
		{"pkg/...", "pkg"},
		{"...", "."},
		{"/abs/path/...", "/abs/path"},
		{"/...", "/"},

		// Unchanged inputs.
		{"./pkg", "./pkg"},
		{"pkg", "pkg"},
		{".", "."},
		{"foo.go", "foo.go"},
		{"a...b", "a...b"},   // not a trailing /... pattern
		{"foo....", "foo...."}, // not a trailing /... pattern
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := stripGoEllipsis(tt.in)
			if got != tt.want {
				t.Errorf("stripGoEllipsis(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestProcessPaths_GoEllipsis(t *testing.T) {
	// "./..." should walk recursively from the working directory; here we use
	// an absolute "<tmp>/..." to verify the same expansion path-agnostically.
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
	files := []string{
		filepath.Join(root, "a.go"),
		filepath.Join(root, "sub", "b.go"),
		filepath.Join(root, "sub", "deep", "c.go"),
	}
	for _, p := range files {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(p, []byte(src), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	if err := processPaths([]string{root + "/..."}, os.Stdout); err != nil {
		t.Fatalf("processPaths: %v", err)
	}

	for _, p := range files {
		disk, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		fmtIdx := strings.Index(string(disk), "\"fmt\"")
		osIdx := strings.Index(string(disk), "\"os\"")
		if fmtIdx < 0 || osIdx < 0 || fmtIdx > osIdx {
			t.Errorf("file %s not sorted (ellipsis expansion broken):\n%s", p, string(disk))
		}
	}
}

func TestProcessPaths_DotEllipsis(t *testing.T) {
	// Bare "./..." should resolve relative to the test's working directory.
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
	target := filepath.Join(root, "x.go")
	if err := os.WriteFile(target, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := processPaths([]string{"./..."}, os.Stdout); err != nil {
		t.Fatalf("processPaths: %v", err)
	}

	disk, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Index(string(disk), "\"fmt\"") > strings.Index(string(disk), "\"os\"") {
		t.Errorf("./... did not process file:\n%s", string(disk))
	}
}

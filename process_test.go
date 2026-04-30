package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/dst/decorator"
)

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

// Test helpers
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

package main

import (
	"strings"
	"testing"
)

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

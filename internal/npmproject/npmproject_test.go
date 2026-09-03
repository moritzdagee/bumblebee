package npmproject

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDirectDepsMergesAllSections(t *testing.T) {
	p := filepath.Join(t.TempDir(), "package.json")
	body := `{"dependencies":{"a":"1"},"devDependencies":{"b":"1"},"optionalDependencies":{"c":"1"},"peerDependencies":{"d":"1"}}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := LoadDirectDeps(p, 0, nil)
	for _, n := range []string{"a", "b", "c", "d"} {
		if !got[n] {
			t.Fatalf("missing %s in %v", n, got)
		}
	}
}

func TestLoadDirectDepsMissingFileIsNil(t *testing.T) {
	if got := LoadDirectDeps(filepath.Join(t.TempDir(), "nope.json"), 0, nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestLoadDirectDepsBadJSONDiags(t *testing.T) {
	p := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(p, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	var warned bool
	if got := LoadDirectDeps(p, 0, func(level, path, msg string) { warned = true }); got != nil || !warned {
		t.Fatalf("expected nil + diag, got %v warned=%v", got, warned)
	}
}

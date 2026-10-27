package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveInside(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "ws")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{filepath.Join(root, "src"), outside, filepath.Join(base, "ws-evil")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	ok := []string{"src/a.go", "src/new/deep/b.go", filepath.Join(root, "src", "c.go")}
	for _, p := range ok {
		if _, err := ResolveInside(root, p); err != nil {
			t.Errorf("%s should be allowed: %v", p, err)
		}
	}
	bad := []string{"../outside/x", "/etc/passwd", "link/x", "link/new.txt", "../ws-evil/a", "src/../../outside/x"}
	for _, p := range bad {
		if _, err := ResolveInside(root, p); err != ErrOutsideWorkspace {
			t.Errorf("%s should be rejected, got %v", p, err)
		}
	}
}

func TestRedact(t *testing.T) {
	cases := map[string]string{
		"ghp_" + strings.Repeat("a", 36):             Redacted,
		"Authorization: Bearer abcdefghijklmnop1234": "Authorization: " + Redacted,
		"API_KEY=hunter2":                            "API_KEY=" + Redacted,
		"access ya29." + strings.Repeat("x", 30):     "access " + Redacted,
		"Opened src/main.go in 12 ms":                "Opened src/main.go in 12 ms",
	}
	for in, want := range cases {
		if got := Redact(in); got != want {
			t.Errorf("Redact(%q) = %q, want %q", in, got, want)
		}
	}
}

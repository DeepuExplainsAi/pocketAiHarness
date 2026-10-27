// Package guard holds the checks that keep agent activity inside the user's workspace.
package guard

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrOutsideWorkspace = errors.New("path is outside the workspace")

// ResolveInside returns candidate as an absolute path inside root. Symlinks on the
// nearest existing ancestor are followed, so a link cannot be used to escape.
func ResolveInside(root, candidate string) (string, error) {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	target := candidate
	if !filepath.IsAbs(target) {
		target = filepath.Join(realRoot, target)
	}
	target = filepath.Clean(target)

	existing, rest := target, ""
	for {
		if real, err := filepath.EvalSymlinks(existing); err == nil {
			target = filepath.Join(real, rest)
			break
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			break
		}
		rest = filepath.Join(filepath.Base(existing), rest)
		existing = parent
	}

	rel, err := filepath.Rel(realRoot, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", ErrOutsideWorkspace
	}
	return target, nil
}

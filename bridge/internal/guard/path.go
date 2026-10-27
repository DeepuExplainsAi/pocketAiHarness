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
		target = realRoot + string(os.PathSeparator) + target // Join would Clean and fold ".." early
	}
	// Clean would fold "link/../x" into "x" before the link is followed, hiding an escape.
	// Resolve ".." only after symlinks, by refusing it in the part that exists.
	if hasDotDot(target) {
		var err error
		if target, err = resolveWithDotDot(target); err != nil {
			return "", ErrOutsideWorkspace
		}
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

func hasDotDot(p string) bool {
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// resolveWithDotDot walks the path element by element so ".." applies to the real directory
// reached through any symlink, as the operating system would.
func resolveWithDotDot(p string) (string, error) {
	cur := string(os.PathSeparator)
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		switch part {
		case "", ".":
			continue
		case "..":
			cur = filepath.Dir(cur)
		default:
			next := filepath.Join(cur, part)
			if real, err := filepath.EvalSymlinks(next); err == nil {
				next = real
			}
			cur = next
		}
	}
	return cur, nil
}

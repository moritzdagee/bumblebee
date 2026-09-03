// Package npmproject reads the parts of a package.json that several
// JavaScript lockfile scanners share. LoadDirectDeps used to live twice
// (yarn and bun, review 2026-09-02); a new dependency section added to one
// copy would have made the two scanners disagree on the same package.json.
package npmproject

import (
	"encoding/json"
	"os"
)

// LoadDirectDeps returns the union of dependencies, devDependencies,
// optionalDependencies and peerDependencies declared in the package.json at
// path, as a set. Callers treat nil as "unknown" and leave DirectDependency
// absent. A non-nil diag is invoked for read or parse failures of a
// package.json that exists; missing or non-regular files are silent.
func LoadDirectDeps(path string, maxSize int64, diag func(level, path, msg string)) map[string]bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	if maxSize > 0 && info.Size() > maxSize {
		if diag != nil {
			diag("warn", path, "skipping direct-dependency resolution: package.json exceeds max file size")
		}
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if diag != nil {
			diag("warn", path, "read package.json for direct-dependency resolution: "+err.Error())
		}
		return nil
	}
	var pj struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
	}
	if err := json.Unmarshal(data, &pj); err != nil {
		if diag != nil {
			diag("warn", path, "parse package.json for direct-dependency resolution: "+err.Error())
		}
		return nil
	}
	out := map[string]bool{}
	for _, m := range []map[string]string{pj.Dependencies, pj.DevDependencies, pj.OptionalDependencies, pj.PeerDependencies} {
		for n := range m {
			out[n] = true
		}
	}
	return out
}

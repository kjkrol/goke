package goke_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// dep_rules_test.go enforces the internal package dependency graph.
// Each entry defines the exact set of non-stdlib imports a package is allowed to have.
// Any import not listed here will cause the test to fail.
//
// Layer 0   iter          (→ uid)
// Layer 1   comp          (→ iter)
// Layer 2   chunk, orch   (→ comp)
// Layer 3   colstore      (→ comp, chunk, iter)
// Layer 4   arch          (→ comp, colstore)
// Layer 5   addr          (→ arch, colstore)
// Layer 6   ent           (→ addr, arch, colstore, comp, iter)
// Layer 7   query         (→ addr, arch, colstore, comp, iter)
// Layer 8   reg           (→ ent, arch, comp, query)
//
// github.com/kjkrol/uid is an external module; listed explicitly per package.

const module = "github.com/kjkrol/goke"
const uidPkg = "github.com/kjkrol/uid"

var depRules = map[string][]string{
	"iter": {
		uidPkg,
	},
	"internal/comp": {
		module + "/iter",
	},
	"internal/chunk": {
		module + "/internal/comp",
		uidPkg,
	},
	"internal/orch": {
		module + "/internal/comp",
		uidPkg,
	},
	"internal/colstore": {
		module + "/internal/comp",
		module + "/internal/chunk",
		module + "/iter",
		uidPkg,
	},
	"internal/arch": {
		module + "/internal/comp",
		module + "/internal/colstore",
		uidPkg,
	},
	"internal/addr": {
		module + "/internal/arch",
		module + "/internal/colstore",
		uidPkg,
	},
	"internal/ent": {
		module + "/internal/addr",
		module + "/internal/arch",
		module + "/internal/colstore",
		module + "/internal/comp",
		module + "/iter",
		uidPkg,
	},
	"internal/query": {
		module + "/internal/addr",
		module + "/internal/arch",
		module + "/internal/colstore",
		module + "/internal/comp",
		module + "/iter",
		uidPkg,
	},
	"internal/reg": {
		module + "/internal/arch",
		module + "/internal/comp",
		module + "/internal/ent",
		module + "/internal/query",
		uidPkg,
	},
}

func TestDependencyRules(t *testing.T) {
	for pkg, allowed := range depRules {
		pkg, allowed := pkg, allowed
		t.Run(pkg, func(t *testing.T) {
			imports := listExternalImports(t, module+"/"+pkg)

			allowedSet := make(map[string]bool, len(allowed))
			for _, a := range allowed {
				allowedSet[a] = true
			}

			for _, imp := range imports {
				if !allowedSet[imp] {
					t.Errorf("forbidden import: %s", imp)
				}
			}
		})
	}
}

func listExternalImports(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	if err != nil {
		t.Fatalf("go list %s: %v", pkg, err)
	}
	var info struct{ Imports []string }
	if err := json.Unmarshal(out, &info); err != nil {
		t.Fatalf("unmarshal go list output: %v", err)
	}
	var result []string
	for _, imp := range info.Imports {
		if isExternalPkg(imp) {
			result = append(result, imp)
		}
	}
	return result
}

func isExternalPkg(pkg string) bool {
	return strings.Contains(strings.SplitN(pkg, "/", 2)[0], ".")
}

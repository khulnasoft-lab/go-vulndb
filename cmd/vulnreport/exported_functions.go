package main

import (
	"fmt"
	"strings"

	"github.com/khulnasoft-lab/go-vulndb/cmd/vulnreport/internal/vulnentries"
	"github.com/khulnasoft-lab/go-vulndb/internal/derrors"
	"github.com/khulnasoft-lab/go-vulndb/internal/osvutils"
	"github.com/khulnasoft-lab/go-vulndb/internal/report"
	"github.com/khulnasoft-lab/go-vulndb/internal/version"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

// exportedFunctions returns a set of vulnerable functions exported
// by a packages from the module.
func exportedFunctions(pkg *packages.Package, m *report.Module) (_ map[string]bool, err error) {
	defer derrors.Wrap(&err, "exportedFunctions(%q)", pkg.PkgPath)

	if pkg.Module != nil {
		v := version.TrimPrefix(pkg.Module.Version)
		affected, err := osvutils.AffectsSemver(report.AffectedRanges(m.Versions), v)
		if err != nil {
			return nil, err
		}
		if !affected {
			return nil, fmt.Errorf("version %s of module %s is not affected by this vuln", v, pkg.Module.Path)
		}
	}

	entries, err := vulnentries.Functions([]*packages.Package{pkg}, m)
	if err != nil {
		return nil, err
	}
	// Return the name of all entry points.
	// Note that "main" and "init" are both possible entries.
	// Both have clear meanings: "main" means that invoking
	// the program is a problem, and "init" means that very likely
	// some global state is altered, and so every exported function
	// is vulnerable. For now, we leave it to consumers to use this
	// information as they wish.
	names := map[string]bool{}
	for _, e := range entries {
		if pkgPath(e) == pkg.PkgPath {
			names[symbolName(e)] = true
		}
	}
	return names, nil
}

func symbolName(fn *ssa.Function) string {
	recv := fn.Signature.Recv()
	if recv == nil {
		return fn.Name()
	}
	recvType := recv.Type().String()
	// Remove package path from type.
	i := strings.LastIndexByte(recvType, '.')
	if i < 0 {
		return recvType + "." + fn.Name()
	}
	return recvType[i+1:] + "." + fn.Name()
}

// pkgPath returns the path of the f's enclosing package, if any.
// Otherwise, returns "".
//
// Copy of github.com/khulnasoft-lab/go-vulndb/vuln/internal/vulncheck/source.go:pkgPath.
func pkgPath(f *ssa.Function) string {
	if f.Package() != nil && f.Package().Pkg != nil {
		return f.Package().Pkg.Path()
	}
	return ""
}

package main

import (
	"context"
	"fmt"

	"github.com/khulnasoft-lab/go-vulndb/internal/cveschema5"
	"github.com/khulnasoft-lab/go-vulndb/internal/ghsa"
	"github.com/khulnasoft-lab/go-vulndb/internal/report"
	"golang.org/x/exp/slices"
)

// addMissingAliases uses the existing aliases in a report to find
// any missing aliases, and adds them to the report.
func addMissingAliases(ctx context.Context, r *report.Report, gc *ghsa.Client) (added int) {
	return r.AddAliases(allAliases(ctx, r.Aliases(), gc))
}

// allAliases returns a list of all aliases associated with the given knownAliases,
// (including the knownAliases themselves).
func allAliases(ctx context.Context, knownAliases []string, gc *ghsa.Client) []string {
	aliasesFor := func(ctx context.Context, alias string) ([]string, error) {
		switch {
		case ghsa.IsGHSA(alias):
			return aliasesForGHSA(ctx, alias, gc)
		case cveschema5.IsCVE(alias):
			return aliasesForCVE(ctx, alias, gc)
		default:
			return nil, fmt.Errorf("unsupported alias %s", alias)
		}
	}
	return aliasesBFS(ctx, knownAliases, aliasesFor)
}

func aliasesBFS(ctx context.Context, knownAliases []string,
	aliasesFor func(ctx context.Context, alias string) ([]string, error)) (all []string) {
	var queue []string
	var seen = make(map[string]bool)
	queue = append(queue, knownAliases...)

	for len(queue) > 0 {
		alias := queue[0]
		queue = queue[1:]

		if seen[alias] {
			continue
		}

		seen[alias] = true
		all = append(all, alias)
		aliases, err := aliasesFor(ctx, alias)
		if err != nil {
			errlog.Printf(err.Error())
			continue
		}
		queue = append(queue, aliases...)
	}

	slices.Sort(all)
	return slices.Compact(all)
}

func aliasesForGHSA(ctx context.Context, alias string, gc *ghsa.Client) (aliases []string, err error) {
	sa, err := gc.FetchGHSA(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("could not fetch GHSA record for %s", alias)
	}
	for _, id := range sa.Identifiers {
		if id.Type == "CVE" || id.Type == "GHSA" {
			aliases = append(aliases, id.Value)
		}
	}
	return aliases, nil
}

func aliasesForCVE(ctx context.Context, cve string, gc *ghsa.Client) (aliases []string, err error) {
	sas, err := gc.ListForCVE(ctx, cve)
	if err != nil {
		return nil, fmt.Errorf("could not find GHSAs for CVE %s", cve)
	}
	for _, sa := range sas {
		aliases = append(aliases, sa.ID)
	}
	return aliases, nil
}

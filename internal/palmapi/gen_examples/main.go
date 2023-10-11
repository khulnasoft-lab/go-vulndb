// Command gen_examples generates and stores examples
// that can be used to create prompts / training inputs for the PaLM API.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/khulnasoft-lab/go-vulndb/internal/genericosv"
	"github.com/khulnasoft-lab/go-vulndb/internal/ghsarepo"
	"github.com/khulnasoft-lab/go-vulndb/internal/gitrepo"
	"github.com/khulnasoft-lab/go-vulndb/internal/palmapi"
	"github.com/khulnasoft-lab/go-vulndb/internal/report"
	"github.com/khulnasoft-lab/go-vulndb/internal/stdlib"
	"golang.org/x/exp/maps"
)

var (
	localGHSA = flag.String("lg", os.Getenv("LOCAL_GHSA_DB"), "path to local GHSA repo, instead of cloning remote")
	outFolder = flag.String("out", filepath.Join("internal", "palmapi"), "folder to write files to")
)

func main() {
	localRepo, err := gitrepo.Open(context.Background(), ".")
	if err != nil {
		log.Fatal(err)
	}
	byIssue, _, err := report.All(localRepo)
	if err != nil {
		log.Fatal(err)
	}
	reports := maps.Values(byIssue)

	var c *ghsarepo.Client
	if *localGHSA != "" {
		repo, err := gitrepo.Open(context.Background(), *localGHSA)
		if err != nil {
			log.Fatal(err)
		}
		c, err = ghsarepo.NewClientFromRepo(repo)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("cloning remote GHSA repo (use -lg=path/to/local/ghsa/repo to speed this up)...")
		c, err = ghsarepo.NewClient()
		if err != nil {
			log.Fatal(err)
		}
	}

	examples, err := toExamples(collectVulns(reports, c))
	if err != nil {
		log.Fatal(err)
	}

	if err := examples.WriteFiles(*outFolder); err != nil {
		log.Fatal(err)
	}

	log.Printf("wrote %d examples to %s\n", len(examples), *outFolder)
}

type vuln struct {
	r    *report.Report
	ghsa *genericosv.Entry
}

// collectVulns collects a list of report-GHSA pairs.
// The list is filtered to include only reports that are more likely
// to make good examples for AI prompts.
func collectVulns(reports []*report.Report, c *ghsarepo.Client) []*vuln {
	var vulns []*vuln
	for _, r := range reports {
		if isFirstParty(r) ||
			r.IsExcluded() ||
			len(r.GHSAs) != 1 ||
			r.CVEMetadata != nil ||
			len(r.Description) < 300 {
			continue
		}

		v := &vuln{r: r}

		for _, ghsa := range r.GHSAs {
			osv := c.ByGHSA(ghsa)
			if osv == nil {
				log.Printf("GHSA %s not found", ghsa)
			}
			v.ghsa = osv
		}

		// Encourage shortening the GHSA description.
		if len(v.ghsa.Details) < len(r.Description) {
			continue
		}

		vulns = append(vulns, v)
	}

	return vulns
}

func isFirstParty(r *report.Report) bool {
	for _, m := range r.Modules {
		if stdlib.IsStdModule(m.Module) || stdlib.IsCmdModule(m.Module) || stdlib.IsXModule(m.Module) {
			return true
		}
	}
	return false
}

func toExamples(vs []*vuln) (palmapi.Examples, error) {
	var es palmapi.Examples
	for _, v := range vs {
		if v.r == nil || v.ghsa == nil {
			return nil, errors.New("invalid example")
		}
		ex := &palmapi.Example{
			Input: palmapi.Input{
				Module:      v.r.Modules[0].Module,
				Description: v.ghsa.Details,
			},
			Suggestion: palmapi.Suggestion{
				Summary:     removeNewlines(v.r.Summary),
				Description: removeNewlines(v.r.Description),
			},
		}
		es = append(es, ex)
	}
	return es, nil
}

func removeNewlines(s string) string {
	newlines := regexp.MustCompile(`\n+`)
	return newlines.ReplaceAllString(strings.TrimSpace(s), " ")
}

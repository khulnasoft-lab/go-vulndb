package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/khulnasoft-lab/go-vulndb/cmd/modinfo/internal/pkgsitedb"
	"golang.org/x/exp/maps"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("missing module paths")
	}
	var cfg pkgsitedb.Config
	cfg.User = mustEnv("MODINFO_USER")
	cfg.Password = mustEnv("MODINFO_PASSWORD")
	cfg.Host = "127.0.0.1"
	cfg.Port = "5429"
	cfg.DBName = mustEnv("MODINFO_DBNAME")
	ctx := context.Background()
	db, err := pkgsitedb.Open(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Could not open DB. You may need to run

    cloud-sql-proxy $MODINFO_DB?port=5429 &

Details: %v`, err)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	defer db.Close()
	for _, modpath := range args {
		mod, err := pkgsitedb.QueryModule(ctx, db, modpath)
		if err != nil {
			log.Fatal(err)
		}
		display(mod)
	}
}

func mustEnv(ev string) string {
	if r := os.Getenv(ev); r != "" {
		return r
	}
	fmt.Fprintf(os.Stderr, "need value for environment variable %s\n", ev)
	os.Exit(1)
	return ""
}

func display(m *pkgsitedb.Module) {
	if len(m.Packages) == 0 {
		fmt.Printf("No packages for module %s; maybe it doesn't exist?\n", m.Path)
		return
	}
	versionMap := map[string]bool{}
	for _, p := range m.Packages {
		versionMap[p.Version] = true
	}
	versions := maps.Keys(versionMap)
	sort.Strings(versions)
	fmt.Printf("==== %s @ %s ====\n", m.Path, strings.Join(versions, ", "))
	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "Import Path", "Importers")
	for _, p := range m.Packages {
		fmt.Fprintf(tw, "%s\t%5d\n", p.Path, p.NumImporters)
	}
	tw.Flush()
}

package proxy

import (
	"errors"
	"flag"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var realProxy = flag.Bool("proxy", false, "if true, contact the real module proxy and update expected responses")

func TestCanonicalModulePath(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name    string
		path    string
		version string
		want    string
	}{
		{
			name:    "non-canonical",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "0.0.0-20230522180520-0cbf4ffdb4e7",
			want:    "github.com/khulnasoft-lab/go-vulndb",
		},
		{
			name:    "canonical",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "0.0.0-20230522180520-0cbf4ffdb4e7",
			want:    "github.com/khulnasoft-lab/go-vulndb",
		},
		{
			name:    "module needs to be escaped",
			path:    "github.com/RobotsAndPencils/go-saml",
			version: "0.0.0-20230606195814-29020529affc",
			want:    "github.com/RobotsAndPencils/go-saml",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.CanonicalModulePath(tc.path, tc.version)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("CanonicalModulePath() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanonicalModuleVersion(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name    string
		path    string
		version string
		want    string
	}{
		{
			name:    "tagged version already canonical",
			path:    "github.com/khulnasoft-lab/go-vulndb/vuln",
			version: "0.1.0",
			want:    "0.1.0",
		},
		{
			name:    "pseudo-version already canonical",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "0.0.0-20230522180520-0cbf4ffdb4e7",
			want:    "0.0.0-20230522180520-0cbf4ffdb4e7",
		},
		{
			name:    "commit hash",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "0cbf4ffdb4e70fce663ec8d59198745b04e7801b",
			want:    "0.0.0-20230522180520-0cbf4ffdb4e7",
		},
		{
			name:    "module needs to be escaped",
			path:    "github.com/RobotsAndPencils/go-saml",
			version: "0.0.0-20230606195814-29020529affc",
			want:    "0.0.0-20230606195814-29020529affc",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.CanonicalModuleVersion(tc.path, tc.version)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("CanonicalModuleVersion() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestModuleExistsAtTaggedVersion(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name    string
		path    string
		version string
		want    bool
	}{
		{
			name:    "exists",
			path:    "github.com/khulnasoft-lab/go-vulndb/vuln",
			version: "0.1.0",
			want:    true,
		},
		{
			name:    "non-canonical module ok",
			path:    "github.com/khulnasoft-lab/go-vulndb/vuln",
			version: "0.1.0",
			want:    true,
		},
		{
			name:    "module needs to be escaped",
			path:    "github.com/Masterminds/squirrel",
			version: "1.5.4",
			want:    true,
		},
		{
			name:    "non-canonical version not OK",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "0cbf4ffdb4e70fce663ec8d59198745b04e7801b",
			want:    false,
		},
		{
			name:    "module exists, version does not",
			path:    "github.com/khulnasoft-lab/go-vulndb",
			version: "1.0.0",
			want:    false,
		},
		{
			name:    "neither exist",
			path:    "golang.org/x/notamod",
			version: "1.0.0",
			want:    false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if got := c.ModuleExistsAtTaggedVersion(tc.path, tc.version); got != tc.want {
				t.Errorf("ModuleExistsAtTaggedVersion() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVersions(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "no tagged versions",
			path: "github.com/khulnasoft-lab/go-vulndb",
			want: nil,
		},
		{
			name: "tagged versions",
			path: "github.com/khulnasoft-lab/go-vulndb/vuln",
			want: []string{
				"0.1.0",
				"0.2.0",
				"1.0.0",
				"1.0.1",
			},
		},
		{
			name: "module needs to be escaped",
			path: "github.com/RobotsAndPencils/go-saml",
			want: nil, // no tagged versions
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.Versions(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Versions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLatest(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		path string
		want string
	}{
		{
			path: "github.com/khulnasoft-lab/go-vulndb",
			want: "0.0.0-20230911193511-c7cbbd05f085",
		},
		{
			path: "github.com/khulnasoft-lab/go-vulndb/vuln",
			want: "1.0.1",
		},
		{
			// module needs to be escaped
			path: "github.com/RobotsAndPencils/go-saml",
			want: "0.0.0-20230606195814-29020529affc",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.path, func(t *testing.T) {
			got, err := c.Latest(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("Latest() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFindModule(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name    string
		path    string
		want    string
		wantErr error
	}{
		{
			name: "module is a prefix of path",
			path: "k8s.io/kubernetes/staging/src/k8s.io/apiserver/pkg/server",
			want: "k8s.io/kubernetes/staging/src/k8s.io/apiserver",
		},
		{
			name: "path is a module",
			path: "k8s.io/kubernetes/staging/src/k8s.io/apiserver",
			want: "k8s.io/kubernetes/staging/src/k8s.io/apiserver",
		},
		{
			name:    "no module",
			path:    "example.co.io/module/package/src/versions/v8",
			wantErr: errNoModuleFound,
		},
		{
			name: "module needs to be escaped",
			path: "github.com/RobotsAndPencils/go-saml/util",
			want: "github.com/RobotsAndPencils/go-saml",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.FindModule(tc.path)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("FindModule() error = %v, want err containing %v", err, tc.wantErr)
			} else if got != tc.want {
				t.Errorf("FindModule() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestModuleExists(t *testing.T) {
	c, err := NewTestClient(t, *realProxy)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "exists",
			path: "k8s.io/kubernetes",
			want: true,
		},
		{
			name: "exists (needs escape)",
			path: "github.com/RobotsAndPencils/go-saml",
			want: true,
		},
		{
			name: "does not exist",
			path: "example.com/not/a/module",
			want: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := c.ModuleExists(tc.path)
			if got != tc.want {
				t.Errorf("ModuleExists() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCacheAndErrors(t *testing.T) {
	okEndpoint, notFoundEndpoint := "endpoint", "not/found"
	okResponse := "response"
	responses := map[string]*response{
		okEndpoint: {
			Body:       okResponse,
			StatusCode: http.StatusOK,
		},
		notFoundEndpoint: {
			Body:       "",
			StatusCode: http.StatusNotFound,
		},
	}
	c, cleanup := fakeClient(responses)
	t.Cleanup(cleanup)

	wantHits := 3
	for i := 0; i < wantHits+1; i++ {
		b, err := c.lookup(okEndpoint)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(b), okResponse; got != want {
			t.Errorf("lookup(%q) = %s, want %s", okEndpoint, got, want)
		}
	}
	if c.cache.hits != wantHits {
		t.Errorf("cache hits = %d, want %d", c.cache.hits, wantHits)
	}

	if _, err := c.lookup(notFoundEndpoint); err == nil {
		t.Errorf("lookup(%q) succeeded, want error", notFoundEndpoint)
	}

	want, got := responses, c.responses()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Responses() unexpected diff (want-, got+):\n%s", diff)
	}
}

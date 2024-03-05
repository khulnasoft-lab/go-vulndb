FROM golang:1.22.1-alpine

# This Dockerfile sets up an image for repeated integration testing.
# This assumes the build context, i.e., CWD is vuln/

# ---- Step 0: Setup shared build tools. ----
RUN apk update && apk add bash git gcc musl-dev linux-headers gcompat

# ---- Step 1: Build govulncheck ----
COPY . /go/src/github.com/khulnasoft-lab/go-vulndb/vuln
WORKDIR /go/src/github.com/khulnasoft-lab/go-vulndb/vuln/cmd/govulncheck/integration
RUN go install github.com/khulnasoft-lab/go-vulndb/vuln/cmd/govulncheck

# ---- Step 2: Build other test binaries ----
RUN go install golang.org/dl/go1.18@latest
RUN go install github.com/khulnasoft-lab/go-vulndb/vuln/cmd/govulncheck/integration/k8s
RUN go install github.com/khulnasoft-lab/go-vulndb/vuln/cmd/govulncheck/integration/stackrox-scanner

VERSION ?= dev

.PHONY: all

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o bin/nodekit .
test:
	go test -coverprofile=coverage.out -coverpkg=./... -covermode=atomic ./...
# Resolve the highest-versioned go-algorand release tagged -stable. Sorted by
# version rather than publish date, so a late backport on an older line (say
# v4.7.5-stable shipped after v5.0.1-stable) cannot win. Pin a specific release
# with ALGOD_VERSION=v5.0.1-stable. Note that specs older than v4.7.4-stable
# predate the uint64 round types nodekit now uses and will not compile.
# Tags are listed with git ls-remote rather than the GitHub REST API, which
# allows 60 unauthenticated requests an hour and is routinely exhausted on
# shared CI runners.
ALGOD_REPO = https://github.com/algorand/go-algorand
# Run the codegen at the version go.mod requires, read from go.mod rather than
# written out here: `go run pkg@version` ignores the module's requirements
# entirely, so a hardcoded version silently drifts from the library the
# generated code is compiled against the first time anyone runs `go get -u`.
# Recursively expanded (=, not :=) so the query only runs for `make generate`.
#
# Do NOT use a bare `oapi-codegen`: go-algorand developers often have the
# algorand/oapi-codegen v1 fork on PATH, which cannot parse this v2-style
# generate.yaml.
OAPI_CODEGEN_MODULE = github.com/oapi-codegen/oapi-codegen/v2
OAPI_CODEGEN_VERSION = $(shell go list -m -f '{{.Version}}' $(OAPI_CODEGEN_MODULE))
generate:
	@version="$${ALGOD_VERSION:-$$(git ls-remote --tags --refs '$(ALGOD_REPO)' 'v*-stable' | sed 's|.*/||' | sort -V | tail -n 1)}"; \
	if [ -z "$$version" ]; then echo "could not resolve the latest stable go-algorand release" >&2; exit 1; fi; \
	codegen="$(OAPI_CODEGEN_VERSION)"; \
	if [ -z "$$codegen" ]; then echo "could not read the $(OAPI_CODEGEN_MODULE) version from go.mod" >&2; exit 1; fi; \
	echo "Generating API client from go-algorand $$version with oapi-codegen $$codegen"; \
	go run $(OAPI_CODEGEN_MODULE)/cmd/oapi-codegen@$$codegen -config generate.yaml "https://raw.githubusercontent.com/algorand/go-algorand/$$version/daemon/algod/api/algod.oas3.yml"

GO          ?= go
GOMOD2NIX   ?= go tool gomod2nix
GOPLS       ?= gopls
GINKGO      ?= go tool ginkgo
GOLANGCI    ?= golangci-lint
GORELEASER  ?= goreleaser

build:
	nix build .# .#mockfs --no-link

test:
	$(GINKGO) -r

cover: coverprofile.out
	$(GO) tool cover -func=coverprofile.out
coverprofile.out: $(shell find . -name '*.go')
	$(GINKGO) -r --cover

clean:
	find . \( -name '*cover*' -o -name 'result*' \) -delete

lint:
	$(GOLANGCI) run ./...

format fmt:
	nix fmt

validate:
	curl --data-binary @codecov.yml https://codecov.io/validate

generate gen: .golangci-lint-version
	cd mockfs && $(GO) generate ./...

snapshot:
	$(GORELEASER) release --snapshot --clean
	cd mockfs && $(GORELEASER) release --snapshot --clean

gomod2nix.toml: export GOWORK := off
gomod2nix.toml: go.mod go.sum
	$(GOMOD2NIX) generate

mockfs/gomod2nix.toml: export GOWORK := off
mockfs/gomod2nix.toml: mockfs/go.mod mockfs/go.sum
	cd mockfs && $(GOMOD2NIX) generate

.PHONY: docs/gopls.instructions.md
docs/gopls.instructions.md:
	$(GOPLS) mcp -instructions > $@

.golangci-lint-version: flake.nix flake.lock
	$(GOLANGCI) version --short > $@

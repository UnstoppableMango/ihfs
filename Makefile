GO        ?= go
GOMOD2NIX ?= go tool gomod2nix
GOPLS     ?= go tool gopls
GINKGO    ?= go tool ginkgo

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

format fmt:
	nix fmt

validate:
	curl --data-binary @codecov.yml https://codecov.io/validate

generate gen:
	cd mockfs && $(GO) generate ./...

gomod2nix.toml: export GOWORK := off
gomod2nix.toml: go.mod go.sum
	$(GOMOD2NIX) generate

mockfs/gomod2nix.toml: export GOWORK := off
mockfs/gomod2nix.toml: mockfs/go.mod mockfs/go.sum
	cd mockfs && $(GOMOD2NIX) generate

.PHONY: docs/gopls.instructions.md
docs/gopls.instructions.md:
	$(GOPLS) mcp -instructions > $@

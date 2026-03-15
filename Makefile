include ./go.mk

.DEFAULT_GOAL := build

build:
	nix build .# .#ctrfs .#ghfs .#mockfs --no-link

clean:
	find . \( -name '*cover*' -o -name 'result*' \) -delete

lint:
	$(GOLANGCI) run ./...

format fmt:
	nix fmt

validate:
	curl --data-binary @codecov.yml https://codecov.io/validate

generate gen:
	$(MAKE) -C mockfs generate

gomod2nix: gomod2nix.toml
	$(MAKE) -C ctrfs gomod2nix.toml
	$(MAKE) -C ghfs gomod2nix.toml
	$(MAKE) -C mockfs gomod2nix.toml

.PHONY: docs/gopls.instructions.md
docs/gopls.instructions.md:
	$(GOPLS) mcp -instructions > $@

.golangci-lint-version: flake.nix flake.lock
	$(GOLANGCI) version --short > $@

.PHONY: ghfs mockfs
ghfs:
	$(MAKE) -C ghfs
mockfs:
	$(MAKE) -C mockfs

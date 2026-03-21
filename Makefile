include ./go.mk

.DEFAULT_GOAL := build

build:
	nix build .# .#ctrfs .#ghfs .#mockfs --no-link

clean:
	find . \( -name '*cover*' -o -name 'result*' \) -delete

lint:
	$(GOLANGCI) run ./...

check: lint validate
	nix flake check

format fmt:
	nix fmt

validate:
	curl -s --data-binary @codecov.yml https://codecov.io/validate | head -n 1

generate gen: docs/gopls.instructions.md
	$(MAKE) -C mockfs generate

gomod2nix: gomod2nix.toml
	$(MAKE) -C ctrfs gomod2nix.toml
	$(MAKE) -C ghfs gomod2nix.toml
	$(MAKE) -C mockfs gomod2nix.toml

update:
	nix flake update

docs/gopls.instructions.md: flake.lock
	$(GOPLS) mcp -instructions > $@

.golangci-lint-version: flake.lock
	$(GOLANGCI) version --short > $@

.PHONY: ghfs mockfs
ghfs:
	$(MAKE) -C ghfs
mockfs:
	$(MAKE) -C mockfs

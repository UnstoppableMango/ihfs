GO        ?= go
GOMOD2NIX ?= go tool gomod2nix
GINKGO    ?= go tool ginkgo

test:
	$(GINKGO) -r

cover: coverprofile.out
	$(GO) tool cover -func=coverprofile.out
coverprofile.out: $(shell find . -name '*.go')
	$(GINKGO) -r --cover

gomod2nix.toml: export GOWORK := off
gomod2nix.toml: go.mod go.sum
	$(GOMOD2NIX)

GO         ?= go
GOMOD2NIX  ?= gomod2nix
GOPLS      ?= gopls
GINKGO     ?= ginkgo
GOLANGCI   ?= golangci-lint
GORELEASER ?= goreleaser

TEST_ARGS += -r

ifeq (${CI},true)
	TEST_ARGS += --github-output --race --trace --randomize-all
endif

test:
	$(GINKGO) ${TEST_ARGS}

cover: coverprofile.out
	$(GO) tool cover -func=coverprofile.out
coverprofile.out: $(shell find . -name '*.go')
	$(GINKGO) ${TEST_ARGS} --cover

gomod2nix.toml: export GOWORK := off
gomod2nix.toml: go.mod go.sum
	$(GOMOD2NIX) generate

snapshot:
	$(GORELEASER) release --snapshot --clean

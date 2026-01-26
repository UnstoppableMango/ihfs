GO        ?= go
GOMOD2NIX ?= go tool gomod2nix
GINKGO    ?= go tool ginkgo

build:
	nix build .#ihfs

test:
	$(GINKGO) -r

cover: coverprofile.out
	$(GO) tool cover -func=coverprofile.out
coverprofile.out: $(shell find . -name '*.go')
	$(GINKGO) -r --cover

clean:
	find . -name '*cover*' -delete

format fmt:
	nix fmt

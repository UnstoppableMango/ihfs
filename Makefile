GO        ?= go
GOMOD2NIX ?= go tool gomod2nix
GINKGO    ?= go tool ginkgo

build:
	nix build .#

test:
	$(GINKGO) -r

clean:
	find . -name '*cover*' -delete

format fmt:
	nix fmt

validate:
	curl --data-binary @codecov.yml https://codecov.io/validate

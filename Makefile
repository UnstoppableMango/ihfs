include ./go.mk

build:
	nix build .#

clean:
	find . -name '*cover*' -delete

format fmt:
	nix fmt

validate:
	curl --data-binary @codecov.yml https://codecov.io/validate

.PHONY: ghfs
ghfs:
	$(MAKE) -C ghfs

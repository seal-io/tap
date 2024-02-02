SHELL := /bin/bash

# Borrowed from https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
curr_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# Borrowed from https://stackoverflow.com/questions/2214575/passing-arguments-to-make-run
rest_args := $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))
$(eval $(rest_args):;@:)

targets := $(shell ls $(curr_dir)/hack | grep '.sh' | sed 's/\.sh//g')
$(targets):
	@$(curr_dir)/hack/$@.sh $(rest_args)

help:
	#
	# Usage:
	#
	#   * [dev] `make deps`, get dependencies.
	#
	#   * [dev] `make lint`, check style.
	#           - `BUILD_TAGS="netgo" make lint` check with specified tags.
	#           - `LINT_DIRTY=true make lint` verify whether the code tree is dirty.
	#
	#   * [dev] `make test`, execute unit testing.
	#           - `BUILD_TAGS="netgo" make test` test with specified tags.
	#
	#   * [dev] `make build`, execute cross building.
	#           - `VERSION=vX.y.z+l.m make build` build all targets with vX.y.z+l.m version.
	#
	#   * [ci]  `make ci`, execute `make deps`, `make lint`, `make test`, `make testacc`, `make build` and `make package`.
	#           - `CI_CHECK=false make ci` only execute `make build` and `make package`.
	#           - `CI_PUBLISH=false make ci` only execute `make deps`, `make lint`, `make test` and `make testacc`.
	#
	#   * [ci]  `make release`, prepare archives checksum and sign checksum with GPG.
	#           - `GPG_FINGERPRINT="..." make release` sign checksum with specified GPG key.
	#
	@echo

.DEFAULT_GOAL := build
.PHONY: $(targets)

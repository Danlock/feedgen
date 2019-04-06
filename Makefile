GOVERSION = $(shell go version)
BUILDTIME = $(shell date -u --rfc-3339=seconds)
GITHASH = $(shell git describe --dirty --always --tags)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = $(GITCOMMITNO).$(GITHASH)
LONGBUILDTAG = $(BUILDTIME).$(GOVERSION)
LDFLAGS = -ldflags "-X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(LONGBUILDTAG)'"
.PHONY: example

gen: design/*
	@rm -rf gen/*
	@goa gen github.com/danlock/go-rss-gen/design

example:
	@rm -rf example/*
	@goa example github.com/danlock/go-rss-gen/design -o example
	@goa gen github.com/danlock/go-rss-gen/design -o example

build:
	go build $(LDFLAGS) -o ./bin/rss_gen ./cmd/rss_gen
	go build -o ./bin/rss_gen_cli ./cmd/rss_gen-cli
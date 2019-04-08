BUILDTIME = $(shell date -u --rfc-3339=seconds)
GITHASH = $(shell git describe --dirty --always --tags)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = $(GITCOMMITNO).$(GITHASH)
LONGBUILDTAG = Build Time:$(BUILDTIME)
LDFLAGS = -X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(LONGBUILDTAG)'
RELEASELDFLAGS = $(LDFLAGS) -X 'github.com/danlock/go-rss-gen/lib/logger.isDebug=f'
.PHONY: example

gen: design/*
	@rm -rf gen/*
	@goa gen github.com/danlock/go-rss-gen/design

example:
	@rm -rf example/*
	@goa example github.com/danlock/go-rss-gen/design -o example
	@goa gen github.com/danlock/go-rss-gen/design -o example

build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/rss_gen ./cmd/rss_gen
	go build -o ./bin/rss_gen_cli ./cmd/rss_gen-cli

release:
	go build -ldflags "$(RELEASELDFLAGS)" -o ./bin/rss_gen ./cmd/rss_gen
	go build -o ./bin/rss_gen_cli ./cmd/rss_gen-cli

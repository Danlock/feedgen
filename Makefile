BUILDTIME = $(shell date -u --rfc-3339=seconds)
GITHASH = $(shell git describe --dirty --always --tags)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = $(GITCOMMITNO).$(GITHASH)
LONGBUILDTAG = Build Time:$(BUILDTIME)
LDFLAGS = -X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(LONGBUILDTAG)'
.PHONY: build

gen: design/*
	@rm -rf gen/*
	@swagger generate server -t gen -A feedgen -f design/api.yml --exclude-main

build:
	go build -race -ldflags "$(LDFLAGS)" -o ./bin/feedgen ./cmd/feedgen

docker-build:
	@docker build -t feedgen .

docker-deploy: docker-build
	@docker save -o /tmp/feedgen feedgen

version: 
	@echo $(SHORTBUILDTAG)
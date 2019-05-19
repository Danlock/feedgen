BUILDTIME = $(shell date -u --rfc-3339=seconds)
GITHASH = $(shell git describe --dirty --always --tags)
GITCOMMITNO = $(shell git rev-list --all --count)
SHORTBUILDTAG = $(GITCOMMITNO).$(GITHASH)
LONGBUILDTAG = Build Time:$(BUILDTIME)
LDFLAGS = -X 'main.buildTag=$(SHORTBUILDTAG)' -X 'main.buildInfo=$(LONGBUILDTAG)'
.PHONY: build

depend:
	@GO111MODULE=on go get -u
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go get tidy

gen: design/*
	@rm -rf gen/*
	@swagger generate server -t gen -A feedgen -f design/api.yml --exclude-main

build: gen
	GO111MODULE=on go build -mod=vendor -race -ldflags "$(LDFLAGS)" -o ./bin/feedgen ./cmd/feedgen

docker-build:
	@docker build -t feedgen .
	@docker save -o /tmp/feedgen feedgen

deploy: docker-build
	@scp /tmp/feedgen root@feedgen.xyz:/tmp/feedgen
	@ssh root@feedgen.xyz "cd ~/feedgen && make restart"

restart:
	@docker load -i /tmp/feedgen
	@cd ops && docker-compose up -d --force-recreate api poll

version: 
	@echo $(SHORTBUILDTAG)
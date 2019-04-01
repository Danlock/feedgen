
.PHONY: example

gen: design/*
	@rm -rf gen/*
	@goa gen github.com/danlock/go-rss-gen/design

example:
	@rm -rf example/*
	@goa example github.com/danlock/go-rss-gen/design -o example
	@goa gen github.com/danlock/go-rss-gen/design -o example
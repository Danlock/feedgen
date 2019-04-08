package api

import (
	"context"

	"github.com/danlock/go-rss-gen/db"

	"github.com/danlock/go-rss-gen/gen/feedgen"
	"github.com/danlock/go-rss-gen/lib/logger"
)

// feedgen service example implementation.
// The example methods log the requests and return zero values.
type fgService struct {
	mangaStore db.MangaStorer
}

// New returns the feedgen service implementation.
func NewFeedSrvc(ms db.MangaStorer) feedgen.Service {
	return &fgService{ms}
}

// Manga implements manga.
func (s *fgService) Manga(ctx context.Context, p *feedgen.MangaPayload) (res string, err error) {
	logger.Infof(ctx, "feedgen.manga")
	return
}

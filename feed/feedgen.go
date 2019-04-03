package feed

import (
	"context"

	"github.com/danlock/go-rss-gen/gen/feedgen"
	"github.com/danlock/go-rss-gen/lib/logger"
)

// feedgen service example implementation.
// The example methods log the requests and return zero values.
type fgService struct {
}

// New returns the feedgen service implementation.
func New() feedgen.Service {
	return &fgService{}
}

// Manga implements manga.
func (s *fgService) Manga(ctx context.Context, p *feedgen.MangaPayload) (res string, err error) {
	logger.Infof(ctx, "feedgen.manga")
	return
}

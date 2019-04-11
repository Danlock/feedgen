package api

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/danlock/go-rss-gen/gen/http/feedgen/client"

	"github.com/danlock/go-rss-gen/lib/logger"
	"github.com/danlock/go-rss-gen/scrape"

	"github.com/danlock/go-rss-gen/db"
	"github.com/gorilla/feeds"

	"github.com/danlock/go-rss-gen/gen/feedgen"
)

// feedgen service example implementation.
// The example methods log the requests and return zero values.
type fgService struct {
	hostURI    string
	mangaStore db.MangaStorer
}

// New returns the feedgen service implementation.
func NewFeedSrvc(host string, ms db.MangaStorer) feedgen.Service {
	return &fgService{host, ms}
}

// Manga implements manga.
func (s *fgService) Manga(ctx context.Context, p *feedgen.MangaPayload) (string, error) {
	seenTitles := make(map[string]struct{})
	normalizedTitles := make([]string, 0, len(p.Titles))
	for _, t := range p.Titles {
		t = strings.ToLower(strings.TrimSpace(t))
		if _, seen := seenTitles[t]; seen {
			continue
		}
		seenTitles[t] = struct{}{}
		normalizedTitles = append(normalizedTitles, t)
	}
	releases := make([]db.DBMangaRelease, 0, len(p.Titles))
	if err := s.mangaStore.FindReleasesByTitles(ctx, p.Titles, &releases); err != nil {
		logger.Errf(ctx, "Failed to find releases for those titles err:%+v", err)
		return "", err
	}
	mangaFeed := feeds.Feed{
		Title:       "MangaUpdates Release Page Feed",
		Description: "This feed has the latest releases for the requested titles from MangaUpdates, if those titles have had a release recent enough to be in the database.",
		Created:     time.Now(),
		Link: &feeds.Link{
			Href: s.hostURI + client.MangaFeedgenPath(),
			Rel:  s.hostURI + client.MangaFeedgenPath(),
		},
	}
	for _, r := range releases {
		mangaFeed.Add(&feeds.Item{
			Id:          strconv.Itoa(r.MUID),
			Title:       r.Title,
			Description: r.Release,
			Created:     r.CreatedAt,
			Author: &feeds.Author{
				Name: r.Translators,
			},
			Link: &feeds.Link{
				Href: scrape.GetMUPageURL(r.MUID),
				Rel:  scrape.GetMUPageURL(r.MUID),
			},
		})
	}
	switch p.FeedType {
	case "json":
		return mangaFeed.ToJSON()
	case "atom":
		return mangaFeed.ToAtom()
	case "rss":
		return mangaFeed.ToRss()
	default:
		return "", errors.New("Unsupported feed type")
	}
}

func (s *fgService) ViewManga(ctx context.Context, p *feedgen.ViewMangaPayload) (res string, err error) {
	return
}

package api

import (
	"context"
	"errors"
	"strconv"
	"strings"

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
	hash, err := s.mangaStore.UpsertFeed(ctx, p)
	if err != nil {
		logger.Errf(ctx, "Failed to upsert feed err:%+v", err)
		return "", feedgen.MakeInternalServerError(errors.New(""))
	}
	return s.hostURI + client.ViewMangaFeedgenPath(hash), nil
}

func (s *fgService) ViewManga(ctx context.Context, p *feedgen.ViewMangaPayload) (string, error) {
	feed := db.MangaFeed{}
	if err := s.mangaStore.GetFeed(ctx, p.Hash, &feed); err != nil {
		return "", err
	}
	releases := make([]db.DBMangaRelease, 0, len(feed.Titles))
	if err := s.mangaStore.FindReleasesByTitles(ctx, feed.Titles, &releases); err != nil {
		logger.Errf(ctx, "Failed to find releases for those titles err:%+v", err)
		return "", err
	}
	if len(releases) == 0 {
		logger.Debugf(ctx, "Found no releases for feed %+v, returning empty feed", feed)
	}
	mangaFeed := feeds.Feed{
		Title:       "Feedgen Manga Releases Feed",
		Description: "This feed has the latest releases for the requested titles from MangaUpdates, if those titles have had a release recent enough to be in the database.",
		Created:     feed.CreatedAt,
		Link: &feeds.Link{
			Href: s.hostURI + client.ViewMangaFeedgenPath(feed.Hash),
			Rel:  s.hostURI + client.ViewMangaFeedgenPath(feed.Hash),
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
	mangaFeed.Sort(func(a, b *feeds.Item) bool { return a.Id < b.Id })
	var (
		result string
		err    error
	)
	switch p.FeedType {
	case "atom":
		result, err = mangaFeed.ToAtom()
	case "rss":
		result, err = mangaFeed.ToRss()
	default:
		return "", errors.New("Unsupported feed type")
	}
	if err != nil {
		logger.Errf(ctx, "Failed creating feed %+v err:%+v", mangaFeed, err)
		return "", feedgen.MakeInternalServerError(errors.New(""))
	}
	return result, nil
}

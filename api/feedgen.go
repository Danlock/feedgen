package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/danlock/feedgen/db"
	"github.com/danlock/feedgen/gen/restapi/operations"
	"github.com/danlock/feedgen/lib"
	"github.com/danlock/feedgen/lib/logger"
	"github.com/danlock/feedgen/scrape"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/feeds"
)

// feedgen service example implementation.
// The example methods log the requests and return zero values.
type FgService struct {
	hostURI    *url.URL
	mangaStore db.MangaStorer
}

// New returns the feedgen service implementation.
func NewFeedSrvc(host *url.URL, ms db.MangaStorer) *FgService {
	return &FgService{host, ms}
}
func (s *FgService) Manga(p operations.FeedgenMangaParams) middleware.Responder {
	ctx := p.HTTPRequest.Context()
	seenTitles := make(map[string]struct{})
	normalizedTitles := make([]string, 0, len(p.MangaRequestBody.Titles))
	for _, t := range p.MangaRequestBody.Titles {
		t = strings.ToLower(strings.TrimSpace(t))
		if _, seen := seenTitles[t]; seen {
			continue
		}
		seenTitles[t] = struct{}{}
		normalizedTitles = append(normalizedTitles, t)
	}
	mangaTitles, err := s.mangaStore.FindMangaByTitlesIntoMangaTitlesSlice(ctx, normalizedTitles)
	if err != nil {
		logger.Errf(ctx, "Failed to upsert feed err:%+v", err)
		return lib.NewResponse(ctx, http.StatusBadGateway)
	}
	// There could possibly be duplicate titles assigned to different manga, that edge case is not being covered
	if len(mangaTitles) < len(normalizedTitles) {
		logger.Warnf(ctx, "Only found %d manga from %d requested titles", len(mangaTitles), len(normalizedTitles))
		notFoundTitles := make([]string, 0, len(normalizedTitles)-len(mangaTitles))
		for _, title := range normalizedTitles {
			found := false
			for _, foundTitle := range mangaTitles {
				if title == foundTitle.OriginalTitle {
					found = true
					break
				}
			}
			if !found {
				notFoundTitles = append(notFoundTitles, title)
			}
		}
		return lib.NewResponse(ctx, http.StatusNotFound).WithMsg(fmt.Sprint(notFoundTitles))
	}
	muids := make([]int, 0, len(mangaTitles))
	for _, t := range mangaTitles {
		muids = append(muids, t.MUID)
	}
	hash, err := s.mangaStore.UpsertFeed(ctx, muids)
	if err != nil {
		logger.Errf(ctx, "Failed to upsert feed err:%+v", err)
		return lib.NewResponse(ctx, http.StatusBadGateway)
	}
	viewMangaBuilder := operations.FeedgenViewMangaURL{Hash: hash}
	viewMangaURL, err := viewMangaBuilder.BuildFull(s.hostURI.Scheme, s.hostURI.Host)
	if err != nil {
		logger.Errf(ctx, "Failed to create view manga url err:%+v", err)
		return lib.NewResponse(ctx, http.StatusInternalServerError)
	}
	return operations.NewFeedgenMangaOK().WithPayload(viewMangaURL.String())
}

func (s *FgService) ViewManga(p operations.FeedgenViewMangaParams) middleware.Responder {
	ctx := p.HTTPRequest.Context()

	feed := db.MangaFeed{}
	if err := s.mangaStore.GetFeed(ctx, p.Hash, &feed); err != nil {
		logger.Errf(ctx, "Failed to get feed releases err:%+v", err)
		return lib.NewResponse(ctx, http.StatusBadGateway)
	}
	releases := make([]db.MangaRelease, 0, len(feed.MUIDs))
	if err := s.mangaStore.FindReleasesForFeed(ctx, feed, &releases); err != nil {
		logger.Errf(ctx, "Failed to find releases for those titles err:%+v", err)
		return lib.NewResponse(ctx, http.StatusBadGateway)
	}
	if len(releases) == 0 {
		logger.Dbgf(ctx, "Found no releases for feed %+v, returning empty feed", feed)
	}
	viewMangaBuilder := operations.FeedgenViewMangaURL{Hash: p.Hash, FeedType: p.FeedType}
	viewMangaURL, err := viewMangaBuilder.BuildFull(s.hostURI.Scheme, s.hostURI.Host)
	if err != nil {
		logger.Errf(ctx, "Failed to create view manga url err:%+v", err)
		return lib.NewResponse(ctx, http.StatusInternalServerError)
	}
	mangaFeed := feeds.Feed{
		Title:       "Feedgen Manga Releases Feed",
		Description: "This feed has the latest releases for the requested titles from MangaUpdates, if those titles have had a release recent enough to be in the database.",
		Created:     feed.CreatedAt,
		Link: &feeds.Link{
			Href: viewMangaURL.String(),
			Rel:  "self",
		},
	}

	for _, r := range releases {
		l := &feeds.Link{
			Href: scrape.GetMUPageURL(r.MUID),
			Rel:  "self",
		}

		it := &feeds.Item{
			Id:          fmt.Sprintf("%s&release=%s", scrape.GetMUPageURL(r.MUID), r.Release),
			Title:       fmt.Sprintf("%s %s", r.Title, r.Release),
			Content:     fmt.Sprintf("%s %s released and translated by %s", r.Title, r.Release, r.Translators),
			Description: fmt.Sprintf("%s %s released and translated by %s", r.Title, r.Release, r.Translators),
			Created:     r.CreatedAt,
			Link:        l,
			Source:      l,
			Author:      &feeds.Author{Name: r.Translators},
		}
		// RSS restricts authors to valid emails only
		if *p.FeedType == "rss" {
			it.Author = nil
			it.Source = nil
		}
		mangaFeed.Add(it)
	}
	mangaFeed.Sort(func(a, b *feeds.Item) bool { return a.Id < b.Id })
	var result string
	switch *p.FeedType {
	case "atom":
		p.HTTPRequest.Header.Set("Accept", "application/xml")
		result, err = mangaFeed.ToAtom()
	case "rss":
		p.HTTPRequest.Header.Set("Accept", "application/xml")
		result, err = mangaFeed.ToRss()
	case "json":
		p.HTTPRequest.Header.Set("Accept", "application/json")
		result, err = mangaFeed.ToJSON()
	default:
		logger.Errf(ctx, "Received unsupported field type %s", *p.FeedType)
		return lib.NewResponse(ctx, http.StatusInternalServerError)
	}
	if err != nil {
		logger.Errf(ctx, "Failed creating feed %+v err:%+v", mangaFeed, err)
		return lib.NewResponse(ctx, http.StatusInternalServerError)
	}

	return operations.NewFeedgenViewMangaOK().WithPayload(result)
}

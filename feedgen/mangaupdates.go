package feedgen

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const mangaUpdateURL = "https://www.mangaupdates.com/releases.html"

type MangaRelease struct {
	Title   string
	Release string
	Group   string
	MULink  string
	MUID    int
}

func parseMangaUpdateDayReleaseHtml(table *html.Node) ([]MangaRelease, error) {
	allMangaReleases := make([]MangaRelease, 0)
	currentMangaRelease := MangaRelease{}

	release := table
	for release.NextSibling != nil {
		release = release.NextSibling
		releaseLinks := htmlquery.FindOne(release, "//a")
		for _, attr := range release.Attr {
			if strings.Contains(attr.Key, "class") {
				switch {
				case strings.Contains(attr.Val, "col-6"):
					if len(currentMangaRelease.Title) > 0 {
						allMangaReleases = append(allMangaReleases, currentMangaRelease)
					}
					currentMangaRelease = MangaRelease{}
					if releaseLinks == nil {
						currentMangaRelease.Title = htmlquery.InnerText(release)
					} else {
						currentMangaRelease.Title = htmlquery.InnerText(releaseLinks)
						for _, a := range releaseLinks.Attr {
							if strings.Contains(a.Key, "href") {
								currentMangaRelease.MULink = a.Val
								query, err := url.ParseQuery(a.Val)
								if err != nil {
									return nil, errors.New("MangaUpdate release has invalid link " + a.Val)
								}
								currentMangaRelease.MUID, err = strconv.Atoi(query.Get("id"))
								if err != nil {
									return nil, errors.New("MangaUpdate release has invalid link id " + a.Val)
								}
							}
						}
					}
				case strings.Contains(attr.Val, "col-4"):
					if releaseLinks == nil {
						currentMangaRelease.Group = htmlquery.InnerText(release)
					} else {
						currentMangaRelease.Group = htmlquery.InnerText(releaseLinks)
					}
				case strings.Contains(attr.Val, "col-2"):
					currentMangaRelease.Release = htmlquery.InnerText(release)
				}
			}
		}
	}
	// The first index has the header row of the table (Title, Releases, etc)
	return allMangaReleases[1:], nil
}

func QueryLast2DaysOfMUReleases() ([]MangaRelease, error) {
	html, err := htmlquery.LoadURL(mangaUpdateURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	todaysReleasesHTML := htmlquery.FindOne(html, "//*[@id=\"main_content\"]/div[2]/div/div[2]/div")
	if todaysReleasesHTML == nil {
		return nil, errors.New("Failed parsing for today releases")
	}
	if todaysReleasesHTML.FirstChild == nil {
		return nil, errors.New("Failed parsing for today releases")
	}
	todaysReleases, err := parseMangaUpdateDayReleaseHtml(todaysReleasesHTML.FirstChild)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	yesterdaysReleasesHTML := htmlquery.FindOne(html, "//*[@id=\"main_content\"]/div[2]/div/div[3]/div")
	if yesterdaysReleasesHTML == nil {
		return nil, errors.New("Failed parsing for yesterdays releases")
	}
	if yesterdaysReleasesHTML.FirstChild == nil {
		return nil, errors.New("Failed parsing for yesterdays releases")
	}
	yesterdaysReleases, err := parseMangaUpdateDayReleaseHtml(yesterdaysReleasesHTML.FirstChild)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return append(todaysReleases, yesterdaysReleases...), nil
}

func PollMUForReleases(ctx context.Context, freq time.Duration) <-chan []MangaRelease {
	out := make(chan []MangaRelease)
	timer := time.NewTicker(freq)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				releases, err := QueryLast2DaysOfMUReleases()
				if err != nil {
					log.Printf("Failed to get releases from MangaUpdates! %+v", err)
				}
				select {
				case <-ctx.Done():
					return
				case out <- releases:
				}
			}
		}
	}()
	return out
}

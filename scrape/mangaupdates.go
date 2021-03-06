package scrape

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/danlock/feedgen/lib"
	"github.com/danlock/feedgen/lib/logger"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const muReleasesURL = "https://www.mangaupdates.com/releases.html"
const muInfoURLFormat = "https://www.mangaupdates.com/series.html?id=%d"

func GetMUPageURL(muid int) string {
	return fmt.Sprintf(muInfoURLFormat, muid)
}

type MangaRelease struct {
	MUID        int
	Title       string
	Release     string
	Translators string
	CreatedAt   time.Time
}

type MangaInfo struct {
	MUID          int
	Titles        []string
	DisplayTitle  string
	LatestRelease string
}

func parseMUDailyReleases(table *html.Node) ([]MangaRelease, error) {
	allMangaReleases := make([]MangaRelease, 0)
	now := time.Now()
	currentMangaRelease := MangaRelease{CreatedAt: now}

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
					currentMangaRelease = MangaRelease{CreatedAt: now}
					if releaseLinks == nil {
						currentMangaRelease.Title = strings.ToLower(htmlquery.InnerText(release))
					} else {
						currentMangaRelease.Title = strings.ToLower(htmlquery.InnerText(releaseLinks))
						for _, a := range releaseLinks.Attr {
							if strings.Contains(a.Key, "href") {
								link, err := url.Parse(a.Val)
								if err != nil {
									return nil, errors.Wrap(err, "MangaUpdate release has invalid link "+a.Val)
								}
								ids := link.Query()["id"]
								if len(ids) > 0 {
									muid, err := strconv.Atoi(ids[0])
									if err == nil {
										currentMangaRelease.MUID = muid
									}
								}
							}
						}
					}
				case strings.Contains(attr.Val, "col-4"):
					if releaseLinks == nil {
						currentMangaRelease.Translators = htmlquery.InnerText(release)
					} else {
						currentMangaRelease.Translators = htmlquery.InnerText(releaseLinks)
					}
				case strings.Contains(attr.Val, "col-2"):
					currentMangaRelease.Release = strings.TrimSpace(htmlquery.InnerText(release))
				}
			}
		}
	}
	// The first index has the header row of the table (Title, Releases, etc)
	return allMangaReleases[1:], nil
}

const ErrInvalidMUID lib.SentinelError = "Invalid MUID"
const maxFailedRequests = 10

func GetAndParseMUMangaPage(ctx context.Context, id int) (m MangaInfo, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(muInfoURLFormat, id), nil)
	if err != nil {
		return m, errors.Wrap(err, "Failed creating request")
	}
	req = req.WithContext(ctx)
	// MangaUpdates doesn't seem to reliably advertise Keep-Alive connection status and Go doesnt handle that very well, so close every request
	req.Close = true
	// Retry requests up to 10 times, waiting a random amount of time between retries up to 30 seconds
	failedRequests := 0
	var resp *http.Response
	for {
		resp, err = http.DefaultClient.Do(req)
		if err == nil {
			break
		}
		if failedRequests >= maxFailedRequests {
			return m, errors.Wrap(err, "Failed sending request")
		}
		failedRequests++
		logger.Errf(ctx, "Failed sending request for id %d %d times. Retrying...", id, failedRequests)
		select {
		case <-time.After(time.Duration(failedRequests) * time.Duration(rand.Intn(3000)) * time.Millisecond):
		case <-ctx.Done():
			return m, ctx.Err()
		}
	}
	root, err := htmlquery.Parse(resp.Body)
	if err != nil {
		resp.Body.Close()
		return m, errors.Wrap(err, "Failed getting page")
	}
	resp.Body.Close()
	seriesInfo := htmlquery.FindOne(root, "/html/body/div[2]/div[2]/div[2]/div[2]/div/div[2]/div[1]")
	if seriesInfo == nil {
		errorInfo := htmlquery.FindOne(root, "/html/body/div[2]/div[2]/div[2]/div[2]/div/div/div/div[2]/div")
		if errorInfo != nil && strings.Contains(htmlquery.InnerText(errorInfo), "You specified an invalid series id.") {
			return m, ErrInvalidMUID
		} else {
			return m, errors.New("Failed to parse info page")
		}
	}
	mainTitleNode := htmlquery.FindOne(seriesInfo, "/div[1]/span[1]")
	if mainTitleNode == nil {
		return m, errors.New("Failed to parse mainTitle")
	}
	mainTitle := htmlquery.InnerText(mainTitleNode)
	strings.TrimSpace(mainTitle)
	if mainTitle == "" {
		return m, errors.New("Got empty title")
	}
	m.MUID = id
	m.DisplayTitle = mainTitle
	m.Titles = append(m.Titles, strings.ToLower(mainTitle))
	assocNamesNode := htmlquery.FindOne(seriesInfo, "/div[3]/div[8]")
	if assocNamesNode == nil || assocNamesNode.FirstChild == nil {
		return m, errors.New("Failed to get associated names")
	}

	assocNameNode := assocNamesNode.FirstChild
	for assocNameNode != nil {
		adTitle := strings.TrimSpace(assocNameNode.Data)
		if adTitle != "N/A" && adTitle != "" && adTitle != "br" {
			m.Titles = append(m.Titles, strings.ToLower(adTitle))
		}
		assocNameNode = assocNameNode.NextSibling
	}
	releasesNode := htmlquery.FindOne(seriesInfo, "/div[3]/div[12]")
	if releasesNode == nil {
		return m, errors.New("Failed to get releases")
	}
	releaseText := htmlquery.InnerText(releasesNode)
	if !strings.Contains(releaseText, "N/A") {
		m.LatestRelease = strings.Split(releaseText, " by ")[0]
	}
	return m, nil
}

func QueryLast2DaysOfMUReleases() ([]MangaRelease, error) {
	html, err := htmlquery.LoadURL(muReleasesURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	todaysReleasesHTML := htmlquery.FindOne(html, "//*[@id=\"main_content\"]/div[2]/div/div[2]/div")
	if todaysReleasesHTML == nil || todaysReleasesHTML.FirstChild == nil {
		return nil, errors.New("Failed parsing for today releases")
	}
	todaysReleases, err := parseMUDailyReleases(todaysReleasesHTML.FirstChild)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	yesterdaysReleasesHTML := htmlquery.FindOne(html, "//*[@id=\"main_content\"]/div[2]/div/div[3]/div")
	if yesterdaysReleasesHTML == nil || yesterdaysReleasesHTML.FirstChild == nil {
		return nil, errors.New("Failed parsing for yesterdays releases")
	}
	yesterdaysReleases, err := parseMUDailyReleases(yesterdaysReleasesHTML.FirstChild)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return append(todaysReleases, yesterdaysReleases...), nil
}

func PollMUForReleases(ctx context.Context, freq time.Duration) <-chan []MangaRelease {
	out := make(chan []MangaRelease)
	timer := time.NewTicker(freq)
	pollFunc := func() {
		start := time.Now()
		releases, err := QueryLast2DaysOfMUReleases()
		if err != nil {
			logger.Errf(ctx, "Failed to get releases from MangaUpdates! %+v", err)
			return
		}
		logger.Dbgf(ctx, "Scraped %d mangaupdates releases in %s", len(releases), time.Since(start).String())
		select {
		case <-ctx.Done():
			return
		case out <- releases:
		}
	}
	go func() {
		defer func() {
			close(out)
			timer.Stop()
		}()
		// Poll immediately on startup instead of waiting for poll duration
		pollFunc()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				pollFunc()
			}
		}
	}()
	return out
}

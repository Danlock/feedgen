package feed

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const muReleasesURL = "https://www.mangaupdates.com/releases.html"
const muInfoURLFormat = "https://www.mangaupdates.com/series.html?id=%d"

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
						currentMangaRelease.Title = htmlquery.InnerText(release)
					} else {
						currentMangaRelease.Title = htmlquery.InnerText(releaseLinks)
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
					currentMangaRelease.Release = htmlquery.InnerText(release)
				}
			}
		}
	}
	// The first index has the header row of the table (Title, Releases, etc)
	return allMangaReleases[1:], nil
}

func getAndParseMUMangaPage(id int) (m MangaInfo, err error) {
	root, err := htmlquery.LoadURL(fmt.Sprintf(muInfoURLFormat, id))
	if err != nil {
		return m, errors.Wrap(err, "Failed getting page")
	}
	seriesInfo := htmlquery.FindOne(root, "/html/body/div[2]/div[2]/div[2]/div[2]/div/div[2]/div[1]")
	if seriesInfo == nil {
		return m, errors.New("Failed to parse info page")
	}
	mainTitleNode := htmlquery.FindOne(seriesInfo, "/div[1]/span[1]")
	if mainTitleNode == nil {
		return m, errors.New("Failed to parse mainTitle")
	}
	mainTitle := htmlquery.InnerText(mainTitleNode)
	if mainTitle == "" {
		return m, errors.New("Got empty title")
	}
	m.MUID = id
	m.Titles = append(m.Titles, mainTitle)
	assocNamesNode := htmlquery.FindOne(seriesInfo, "/div[3]/div[8]")
	if assocNamesNode == nil || assocNamesNode.FirstChild == nil {
		return m, errors.New("Failed to get associated names")
	}

	assocNameNode := assocNamesNode.FirstChild
	for assocNameNode != nil {
		if assocNameNode.Data != "N/A" && assocNameNode.Data != "" && assocNameNode.Data != "br" {
			m.Titles = append(m.Titles, assocNameNode.Data)
		}
		assocNameNode = assocNameNode.NextSibling
	}
	releasesNode := htmlquery.FindOne(seriesInfo, "/div[3]/div[12]")
	if releasesNode == nil || releasesNode.FirstChild == nil || releasesNode.FirstChild.NextSibling == nil || releasesNode.FirstChild.NextSibling.NextSibling == nil {
		return m, errors.New("Failed to get releases")
	}
	m.LatestRelease = strings.Split(htmlquery.InnerText(releasesNode), " by ")[0]
	return m, nil
}

const maxSeriesID = 200000

func QueryAllMUSeries() ([]MangaInfo, error) {
	// Create sender goroutine that sends down ids and closes when done
	idChan := make(chan int)
	go func() {
		for i := 1; i < maxSeriesID; i++ {
			idChan <- i
		}
		close(idChan)
	}()
	// Create worker goroutines that do work on each id and end on chan close
	infoChan := make(chan MangaInfo)
	const maxGoroutines = 10
	wg := &sync.WaitGroup{}
	wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for id := range idChan {
				info, err := getAndParseMUMangaPage(id)
				if err != nil {
					log.Printf("Failure on %d err:%+v", id, err)
					continue
				}
				infoChan <- info
			}
			wg.Done()
		}()
	}
	// Create closer goroutine that waits for workers to complete then closes infoChan
	go func() {
		wg.Wait()
		close(infoChan)
	}()
	// Append info from the chan until its closed
	infos := make([]MangaInfo, 0)
	for info := range infoChan {
		if info.MUID > 0 {
			infos = append(infos, info)
		}
	}
	return infos, nil
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
	go func() {
		defer func() {
			close(out)
			timer.Stop()
		}()
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

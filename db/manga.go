package db

import (
	"context"
	"fmt"

	"github.com/danlock/go-rss-gen/lib/logger"
	"github.com/pkg/errors"

	"github.com/danlock/go-rss-gen/feed"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type MangaStorer interface {
	FindMangaByTitlesIntoMangaTitlesSlice(context.Context, []string) ([]MangaTitle, error)
	FindMangaByTitles(context.Context, []string, interface{}) error
	UpsertManga(context.Context, []feed.MangaInfo) error
	UpsertRelease(context.Context, []feed.MangaRelease) error
}

type mangaStore struct {
	db *sqlx.DB
}

func NewMangaStore(db *sqlx.DB) MangaStorer {
	return &mangaStore{db}
}

func (m *mangaStore) UpsertManga(ctx context.Context, manga []feed.MangaInfo) error {
	mangaQuery := `
	INSERT INTO manga (muid, latest_release)
		%s
		ON CONFLICT (muid)
		DO UPDATE SET latest_release = excluded.latest_release;
	`
	titleQuery := "UPSERT INTO mangatitle (muid,title) %s;"

	muidReleaseArray := make([]interface{}, 0, len(manga)*2)
	mangaValues := "VALUES"
	muidTitleArray := make([]interface{}, 0)
	titleValues := "VALUES"

	for _, m := range manga {
		if m.MUID < 1 {
			logger.Errf(ctx, "Skipping upserting corrupted manga %+v", m)
			continue
		}
		mangaValues += fmt.Sprintf(" (?,?),")
		muidReleaseArray = append(muidReleaseArray, m.MUID, m.LatestRelease)
		for _, t := range m.Titles {
			titleValues += fmt.Sprintf(" (?,?),")
			muidTitleArray = append(muidTitleArray, m.MUID, t)
		}
	}
	// Trim off trailing commas
	mangaValues = mangaValues[:len(mangaValues)-1]
	titleValues = titleValues[:len(titleValues)-1]

	mangaQuery = fmt.Sprintf(mangaQuery, mangaValues)
	if _, err := m.db.ExecContext(ctx, mangaQuery, muidReleaseArray...); err != nil {
		logger.Errf(ctx, "Failed upserting manga with %s\n with error %s", mangaQuery, ErrDetails(err))
		return errors.WithStack(err)
	}

	titleQuery = fmt.Sprintf(titleQuery, titleValues)
	if _, err := m.db.ExecContext(ctx, titleQuery, muidTitleArray...); err != nil {
		logger.Errf(ctx, "Failed upserting titles with %s\n with error %s", titleQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

func (m *mangaStore) UpsertRelease(ctx context.Context, releases []feed.MangaRelease) error {
	releaseQuery := `
	INSERT INTO mangarelease (muid, release, translators)
		%s
	ON CONFLICT (muid,release,translators)
	DO NOTHING;
	`
	releaseValues := "VALUES"
	releaseTitlesWithMissingMUID := make([]string, 0)
	releasesWithMissingMUID := make([]feed.MangaRelease, 0)
	valuesArr := make([]interface{}, 0, len(releases)*3)
	for _, r := range releases {
		if r.MUID < 1 {
			releaseTitlesWithMissingMUID = append(releaseTitlesWithMissingMUID, r.Title)
			releasesWithMissingMUID = append(releasesWithMissingMUID, r)
			continue
		}
		valuesArr = append(valuesArr, r.MUID, r.Release, r.Translators)
		releaseValues += " (?,?,?),"
	}
	foundReleases, err := m.FindMangaByTitlesIntoMangaTitlesSlice(ctx, releaseTitlesWithMissingMUID)
	if err != nil {
		return err
	}
	if len(foundReleases) != len(releaseTitlesWithMissingMUID) {
		logger.Errf(ctx, "Could not find all MUID for all releases, titles missing muids: %+v found releases: %+v", releaseTitlesWithMissingMUID, foundReleases)
	}
	for _, r := range releasesWithMissingMUID {
		for _, m := range foundReleases {
			if r.Title == m.Title {
				valuesArr = append(valuesArr, m.MUID, r.Release, r.Translators)
				releaseValues += " (?,?,?),"
			}
		}
	}
	releaseQuery = fmt.Sprintf(releaseQuery, releaseValues)
	if _, err := m.db.ExecContext(ctx, releaseQuery, valuesArr...); err != nil {
		logger.Errf(ctx, "Failed to get upsert release with query %s and err: %+v", releaseQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

type MangaTitle struct {
	MUID  int
	Title string
}

func (m *mangaStore) FindMangaByTitlesIntoMangaTitlesSlice(ctx context.Context, titles []string) ([]MangaTitle, error) {
	manga := make([]MangaTitle, 0, len(titles))
	if err := m.FindMangaByTitles(ctx, titles, &manga); err != nil {
		return nil, err
	}
	return manga, nil
}
func (m *mangaStore) FindMangaByTitles(ctx context.Context, titles []string, outPtr interface{}) error {
	titleQuery := `
	SELECT muid,title FROM mangatitle WHERE title IN (%s);
	`
	titleValues := ""
	titlesArr := make([]interface{}, 0, len(titles))
	for _, t := range titles {
		titleValues += "'?',"
		titlesArr = append(titlesArr, t)
	}
	titleValues = titleValues[:len(titleValues)-1]
	titleQuery = fmt.Sprintf(titleQuery, titleValues)
	if err := m.db.SelectContext(ctx, outPtr, titleQuery, titlesArr...); err != nil {
		logger.Errf(ctx, "Failed getting titles with %s err:%+v", titleQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

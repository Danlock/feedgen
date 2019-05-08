package db

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/danlock/feedgen/gen/models"

	"github.com/danlock/feedgen/lib/logger"
	"github.com/danlock/feedgen/scrape"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type MangaStorer interface {
	FindMangaByTitlesIntoMangaTitlesSlice(context.Context, []string) ([]MangaTitle, error)
	FindMangaByTitles(context.Context, []string, interface{}) error
	FindReleasesByTitles(context.Context, []string, interface{}) error
	UpsertManga(context.Context, []scrape.MangaInfo) error
	UpsertRelease(context.Context, []scrape.MangaRelease) error
	FilterOutReleasesWithoutMangaInDB(context.Context, []scrape.MangaRelease) ([]scrape.MangaRelease, error)
	UpsertFeed(context.Context, *models.FeedgenMangaRequestBody) (string, error)
	GetFeed(context.Context, string, interface{}) error
}

type mangaStore struct {
	db *sqlx.DB
}

func NewMangaStore(db *sqlx.DB) MangaStorer {
	return &mangaStore{db}
}

func (m *mangaStore) UpsertManga(ctx context.Context, manga []scrape.MangaInfo) error {
	mangaQuery := `INSERT INTO manga (muid, latest_release, display_title) VALUES
%s
ON CONFLICT (muid)
DO NOTHING;`
	titleQuery := "UPSERT INTO mangatitle (muid,title) VALUES %s;"

	muidReleaseArray := make([]interface{}, 0, len(manga)*2)
	mangaValues := ""
	muidTitleArray := make([]interface{}, 0)
	titleValues := ""
	seenMUID := make(map[int]struct{})
	for _, m := range manga {
		if _, seen := seenMUID[m.MUID]; seen || m.MUID < 1 {
			continue
		}
		seenMUID[m.MUID] = struct{}{}
		mangaValues += fmt.Sprintf(" (?,?,?),")
		muidReleaseArray = append(muidReleaseArray, m.MUID, m.LatestRelease, m.DisplayTitle)
		for _, t := range m.Titles {
			titleValues += fmt.Sprintf(" (?,?),")
			muidTitleArray = append(muidTitleArray, m.MUID, strings.TrimSpace(strings.ToLower(t)))
		}
	}
	// Trim off trailing commas
	mangaValues = mangaValues[:len(mangaValues)-1]
	titleValues = titleValues[:len(titleValues)-1]

	mangaQuery = fmt.Sprintf(mangaQuery, mangaValues)
	mangaQuery = m.db.Rebind(mangaQuery)
	if res, err := m.db.ExecContext(ctx, mangaQuery, muidReleaseArray...); err != nil {
		logger.Errf(ctx, "Failed upserting manga with %s\n with error %s", mangaQuery, ErrDetails(err))
		return errors.WithStack(err)
	} else {
		num, _ := res.RowsAffected()
		if num > 0 {
			logger.Infof(ctx, "Upserted %d rows of manga", num)
		}
	}
	titleQuery = fmt.Sprintf(titleQuery, titleValues)
	titleQuery = m.db.Rebind(titleQuery)
	if _, err := m.db.ExecContext(ctx, titleQuery, muidTitleArray...); err != nil {
		logger.Errf(ctx, "Failed upserting titles with %s\n with error %s", titleQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

func (m *mangaStore) UpsertRelease(ctx context.Context, releases []scrape.MangaRelease) error {
	releaseQuery := `
	INSERT INTO mangarelease (muid, release, translators)
		%s
	ON CONFLICT (muid,release,translators)
	DO NOTHING;
	`
	releaseValues := "VALUES"
	valuesArr := make([]interface{}, 0, len(releases)*3)
	releaesMissingMUIDs := 0
	seenMUID := make(map[int]struct{})
	for _, r := range releases {
		if _, seen := seenMUID[r.MUID]; seen || r.MUID < 1 {
			releaesMissingMUIDs++
			continue
		}
		valuesArr = append(valuesArr, r.MUID, r.Release, r.Translators)
		releaseValues += " (?,?,?),"
		seenMUID[r.MUID] = struct{}{}

	}
	if releaesMissingMUIDs > 0 {
		logger.Dbgf(ctx, "Skipping %d releases missing MUIDs", releaesMissingMUIDs)
	}
	logger.Dbgf(ctx, "Preparing to upsert %d releases", len(seenMUID))

	releaseValues = releaseValues[:len(releaseValues)-1]
	releaseQuery = fmt.Sprintf(releaseQuery, releaseValues)
	releaseQuery = m.db.Rebind(releaseQuery)
	if _, err := m.db.ExecContext(ctx, releaseQuery, valuesArr...); err != nil {
		logger.Errf(ctx, "Failed to upsert release with query %s and err: %+v", releaseQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	logger.Dbgf(ctx, "Upserted %d releases", len(valuesArr)/3)
	return nil
}

type MangaTitle struct {
	MUID          int
	OriginalTitle string `db:"title"`
	DisplayTitle  string `db:"display_title"`
}

func (m *mangaStore) FindMangaByTitlesIntoMangaTitlesSlice(ctx context.Context, titles []string) ([]MangaTitle, error) {
	manga := make([]MangaTitle, 0, len(titles))
	if err := m.FindMangaByTitles(ctx, titles, &manga); err != nil {
		return nil, err
	}
	return manga, nil
}
func (m *mangaStore) FindMangaByTitles(ctx context.Context, titles []string, outPtr interface{}) error {
	titleQueryRaw := `
	SELECT manga.muid,mangatitle.title,manga.display_title
	FROM mangatitle
	INNER JOIN manga ON manga.muid=mangatitle.muid
	WHERE title IN (?);
	`
	titleQuery, args, err := sqlx.In(titleQueryRaw, titles)
	if err != nil {
		return errors.Wrap(err, "Failed creating IN query")
	}
	titleQuery = m.db.Rebind(titleQuery)
	if err := m.db.SelectContext(ctx, outPtr, titleQuery, args...); err != nil {
		logger.Errf(ctx, "Failed getting titles with %s err:%+v", titleQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

type DBMangaRelease struct {
	MUID        int
	Title       string `db:"display_title"`
	Release     string
	Translators string
	CreatedAt   time.Time `db:"created_at"`
}

func (m *mangaStore) FindReleasesByTitles(ctx context.Context, titles []string, outPtr interface{}) error {
	releaseQueryRaw := `
	SELECT mangarelease.muid, mangarelease.release, mangarelease.translators, mangarelease.created_at, manga.display_title
		FROM mangarelease
		INNER JOIN mangatitle ON mangatitle.muid=mangarelease.muid
		INNER JOIN manga ON mangatitle.muid=manga.muid
		INNER JOIN (
			SELECT muid, max(mangarelease.created_at) most_recent
					FROM mangarelease
					GROUP BY muid
		) mr ON manga.muid = mr.muid AND mangarelease.created_at = mr.most_recent
	WHERE mangatitle.title IN (?);`
	titleVals := make([]interface{}, 0, len(titles))
	for _, t := range titles {
		titleVals = append(titleVals, strings.TrimSpace(strings.ToLower(t)))
	}
	releaseQuery, args, err := sqlx.In(releaseQueryRaw, titleVals)
	if err != nil {
		return errors.Wrap(err, "Failed creating IN query")
	}
	releaseQuery = m.db.Rebind(releaseQuery)
	if err := m.db.SelectContext(ctx, outPtr, releaseQuery, args...); err != nil {
		logger.Errf(ctx, "Failed to find manga releases by titles with %s err: %s", releaseQuery, ErrDetails(err))
		return errors.WithStack(err)
	}
	return nil
}

func (m *mangaStore) FilterOutReleasesWithoutMangaInDB(ctx context.Context, releases []scrape.MangaRelease) ([]scrape.MangaRelease, error) {
	MUIDs := make([]interface{}, 0, len(releases))
	releasesMissingMUIDs := 0
	for _, r := range releases {
		if r.MUID > 0 {
			MUIDs = append(MUIDs, r.MUID)
		} else {
			releasesMissingMUIDs++
		}
	}
	muidQuery, args, err := sqlx.In("SELECT muid FROM manga WHERE muid IN (?);", MUIDs)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating IN query")
	}
	foundMUIDs := make([]int, 0, len(releases))
	muidQuery = m.db.Rebind(muidQuery)
	if err := m.db.Select(&foundMUIDs, muidQuery, args...); err != nil {
		logger.Errf(ctx, "Failed to find manga from muids with %s err: %s", muidQuery, ErrDetails(err))
		return nil, errors.WithStack(err)
	}
	// If we found all of the releases then we're done
	if len(foundMUIDs) == len(releases) {
		return nil, nil
	}
	newReleases := make([]scrape.MangaRelease, 0)
	for _, r := range releases {
		if r.MUID == 0 {
			continue
		}
		found := false
		for _, id := range foundMUIDs {
			if r.MUID == id {
				found = true
				break
			}
		}
		if !found {
			newReleases = append(newReleases, r)
		}
	}
	return newReleases, nil
}

type MangaFeed struct {
	Hash      string
	Type      string
	Titles    pq.StringArray
	CreatedAt time.Time `db:"created_at"`
}

func (m *mangaStore) UpsertFeed(ctx context.Context, p *models.FeedgenMangaRequestBody) (string, error) {
	h := sha256.New()
	for _, t := range p.Titles {
		h.Write([]byte(t))
	}
	mf := MangaFeed{
		Hash:      base64.RawURLEncoding.EncodeToString(h.Sum(nil)),
		Titles:    p.Titles,
		CreatedAt: time.Now(),
	}
	query := `
	INSERT INTO mangafeed (hash, titles)
	VALUES	(?,?)
	ON CONFLICT (hash)
	DO NOTHING;
`
	query = m.db.Rebind(query)
	_, err := m.db.ExecContext(ctx, query, mf.Hash, mf.Titles)
	if err != nil {
		logger.Errf(ctx, "Failed to upsert feeds with %s err: %s", query, ErrDetails(err))
		return "", errors.WithStack(err)
	}
	return mf.Hash, nil
}

func (m *mangaStore) GetFeed(ctx context.Context, hash string, outPtr interface{}) error {
	query := `
	SELECT titles,created_at FROM mangafeed WHERE hash=?;
	`
	query = m.db.Rebind(query)
	if err := m.db.GetContext(ctx, outPtr, query, hash); err != nil {
		logger.Errf(ctx, "Failed to get feed with %s err %s", query, ErrDetails(err))
	}
	return nil
}

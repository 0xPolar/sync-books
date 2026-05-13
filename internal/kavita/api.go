package kavita

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------
// Account
// ---------------------------------------------------------------------------

// Me returns the current authenticated user.
func (c *Client) Me(ctx context.Context) (*UserDto, error) {
	var u UserDto
	if err := c.do(ctx, "GET", "/api/Account", nil, &u, true); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.userID = u.ID
	c.mu.Unlock()
	return &u, nil
}

// ---------------------------------------------------------------------------
// Libraries
// ---------------------------------------------------------------------------

// Libraries returns every library the authenticated user can see.
func (c *Client) Libraries(ctx context.Context) ([]LibraryDto, error) {
	var out []LibraryDto
	if err := c.do(ctx, "GET", "/api/Library/libraries", nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Series — currently reading / on-deck / detail
// ---------------------------------------------------------------------------

// CurrentlyReading returns the series the user has in-progress.
// `userID` may be 0 to mean "the authenticated user" (their profile must be
// shared if requesting someone else's).
//
// Note: this endpoint returns series, not individual chapters. For per-book
// data (ISBN, per-book cover, per-book authors) call Volumes for each series
// and walk the chapters.
func (c *Client) CurrentlyReading(ctx context.Context, userID, pageNumber, pageSize int) ([]SeriesDto, error) {
	q := url.Values{}
	if userID == 0 {
		c.mu.RLock()
		userID = c.userID
		c.mu.RUnlock()
	}
	if userID > 0 {
		q.Set("userId", strconv.Itoa(userID))
	}
	if pageNumber > 0 {
		q.Set("PageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		q.Set("PageSize", strconv.Itoa(pageSize))
	}
	path := "/api/Series/currently-reading"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out []SeriesDto
	if err := c.do(ctx, "GET", path, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// OnDeck returns series the user should pick up next. Pass libraryID=0 for
// all libraries.
func (c *Client) OnDeck(ctx context.Context, libraryID, pageNumber, pageSize int) ([]SeriesDto, error) {
	q := url.Values{}
	q.Set("libraryId", strconv.Itoa(libraryID))
	if pageNumber > 0 {
		q.Set("PageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		q.Set("PageSize", strconv.Itoa(pageSize))
	}
	var out []SeriesDto
	if err := c.do(ctx, "POST", "/api/Series/on-deck?"+q.Encode(), nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// Series fetches a single series by ID.
func (c *Client) Series(ctx context.Context, seriesID int) (*SeriesDto, error) {
	var s SeriesDto
	path := fmt.Sprintf("/api/Series/%d", seriesID)
	if err := c.do(ctx, "GET", path, nil, &s, true); err != nil {
		return nil, err
	}
	return &s, nil
}

// Volumes returns all volumes (with their chapters) for a series. This is the
// endpoint that surfaces per-chapter ISBN/authors/cover detail.
func (c *Client) Volumes(ctx context.Context, seriesID int) ([]VolumeDto, error) {
	var out []VolumeDto
	path := fmt.Sprintf("/api/Series/volumes?seriesId=%d", seriesID)
	if err := c.do(ctx, "GET", path, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// Chapter fetches a single chapter with full metadata (ISBN, writers, etc.).
func (c *Client) Chapter(ctx context.Context, chapterID int) (*ChapterDto, error) {
	var ch ChapterDto
	path := fmt.Sprintf("/api/Chapter?chapterId=%d", chapterID)
	if err := c.do(ctx, "GET", path, nil, &ch, true); err != nil {
		return nil, err
	}
	return &ch, nil
}

// ---------------------------------------------------------------------------
// Reading progress
// ---------------------------------------------------------------------------

// ChapterProgress returns the user's read position on a chapter.
func (c *Client) ChapterProgress(ctx context.Context, chapterID int) (*ProgressDto, error) {
	var p ProgressDto
	path := fmt.Sprintf("/api/Reader/get-progress?chapterId=%d", chapterID)
	if err := c.do(ctx, "GET", path, nil, &p, true); err != nil {
		return nil, err
	}
	return &p, nil
}

// HasProgress reports whether the user has any reading progress on a series.
func (c *Client) HasProgress(ctx context.Context, seriesID int) (bool, error) {
	var v bool
	path := fmt.Sprintf("/api/Reader/has-progress?seriesId=%d", seriesID)
	if err := c.do(ctx, "GET", path, nil, &v, true); err != nil {
		return false, err
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// Sidecar conveniences
// ---------------------------------------------------------------------------

// BookSummary is a flat, JSON-friendly view of a single readable item (a
// chapter in image libraries, a book in epub/pdf libraries). This is the
// shape your sidecar will most likely emit.
type BookSummary struct {
	SeriesID    int       `json:"series_id"`
	ChapterID   int       `json:"chapter_id"`
	VolumeID    int       `json:"volume_id"`
	LibraryID   int       `json:"library_id"`
	Title       string    `json:"title"`
	SeriesName  string    `json:"series_name"`
	Authors     []string  `json:"authors,omitempty"`
	ISBN        string    `json:"isbn,omitempty"`
	Thumbnail   string    `json:"thumbnail"`
	Pages       int       `json:"pages"`
	PagesRead   int       `json:"pages_read"`
	ProgressPct float64   `json:"progress_pct"`
	LastReadUTC time.Time `json:"last_read_utc,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	WordCount   int64     `json:"word_count,omitempty"`
	Language    string    `json:"language,omitempty"`
	Blacklisted bool      `json:"blacklisted,omitempty"`
}

// ToBookSummary flattens a Chapter (within a known Series) into a sidecar-
// friendly BookSummary. Thumbnail prefers the chapter-level cover (per-book
// for epubs/pdfs) and falls back to the series cover.
func (c *Client) ToBookSummary(s SeriesDto, ch ChapterDto) BookSummary {
	authors := make([]string, 0, len(ch.Writers))
	for _, w := range ch.Writers {
		authors = append(authors, w.Name)
	}

	title := ch.TitleName
	if title == "" {
		title = ch.Title
	}
	if title == "" {
		title = s.Name
	}

	thumb := c.ChapterCoverURL(ch.ID)
	if ch.CoverImage == "" && s.CoverImage != "" {
		thumb = c.SeriesCoverURL(s.ID)
	}

	var pct float64
	if ch.Pages > 0 {
		pct = float64(ch.PagesRead) / float64(ch.Pages) * 100
	}

	return BookSummary{
		SeriesID:    s.ID,
		ChapterID:   ch.ID,
		VolumeID:    ch.VolumeID,
		LibraryID:   s.LibraryID,
		Title:       title,
		SeriesName:  s.Name,
		Authors:     authors,
		ISBN:        ch.ISBN,
		Thumbnail:   thumb,
		Pages:       ch.Pages,
		PagesRead:   ch.PagesRead,
		ProgressPct: pct,
		LastReadUTC: ch.LastReadingProgressUtc.Time,
		Summary:     ch.Summary,
		ReleaseDate: ch.ReleaseDate.Time,
		WordCount:   ch.WordCount,
		Language:    ch.Language,
	}
}

// CurrentlyReadingBooks walks CurrentlyReading → Volumes → Chapters and
// returns a flat list of BookSummary entries for every in-progress chapter.
//
// `minProgressPct` filters out chapters below the given read percentage
// (e.g. pass 40 for "more than 40% read"). Use 0 for no minimum.
//
// `since` filters out chapters last read before that timestamp (e.g. pass
// time.Now().AddDate(0,0,-14) for "read in the last 2 weeks"). Pass the
// zero time for no time filter.
func (c *Client) CurrentlyReadingBooks(ctx context.Context, userID int, minProgressPct float64, since time.Time) ([]BookSummary, error) {
	series, err := c.CurrentlyReading(ctx, userID, 0, 0)
	if err != nil {
		return nil, err
	}

	var out []BookSummary
	for _, s := range series {
		volumes, err := c.Volumes(ctx, s.ID)
		if err != nil {
			return nil, fmt.Errorf("volumes for series %d: %w", s.ID, err)
		}
		for _, v := range volumes {
			for _, ch := range v.Chapters {
				// Only include chapters with progress.
				if ch.PagesRead == 0 {
					continue
				}
				// Volumes endpoint returns lightweight chapters; fetch full
				// metadata (ISBN, writers, etc.) from the Chapter endpoint.
				full, err := c.Chapter(ctx, ch.ID)
				if err != nil {
					return nil, fmt.Errorf("chapter %d: %w", ch.ID, err)
				}
				// Preserve progress fields that come from the volume response.
				full.PagesRead = ch.PagesRead
				full.LastReadingProgressUtc = ch.LastReadingProgressUtc
				full.LastReadingProgress = ch.LastReadingProgress
				bs := c.ToBookSummary(s, *full)
				if bs.ProgressPct < minProgressPct {
					continue
				}
				if !since.IsZero() && bs.LastReadUTC.Before(since) {
					continue
				}
				out = append(out, bs)
			}
		}
	}
	return out, nil
}

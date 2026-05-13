package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-currently-reading/internal/kavita"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	connStr string
	conn    *sql.DB
}

func NewDB(destination string) (*DB, error) {
	db := &DB{conn: nil, connStr: destination}

	if err := db.Init(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Init() error {
	var err error
	db.conn, err = sql.Open("sqlite3", db.connStr)
	if err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS books (
		chapter_id     INTEGER PRIMARY KEY,
		series_id      INTEGER NOT NULL,
		volume_id      INTEGER NOT NULL,
		library_id     INTEGER NOT NULL,
		title          TEXT NOT NULL,
		series_name    TEXT NOT NULL,
		authors        TEXT NOT NULL DEFAULT '[]',
		isbn           TEXT,
		thumbnail      TEXT,
		pages          INTEGER NOT NULL DEFAULT 0,
		pages_read     INTEGER NOT NULL DEFAULT 0,
		progress_pct   REAL    NOT NULL DEFAULT 0,
		last_read_utc  TEXT,
		summary        TEXT,
		release_date   TEXT,
		word_count     INTEGER NOT NULL DEFAULT 0,
		language       TEXT,
		updated_at     TEXT NOT NULL,
		blacklisted    INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_books_series ON books(series_id);
	CREATE INDEX IF NOT EXISTS idx_books_last_read ON books(last_read_utc DESC);
	`
	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

func (db *DB) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

// UpsertBook inserts or replaces a single BookSummary row.
func (db *DB) UpsertBook(b kavita.BookSummary) error {
	return upsertBook(db.conn, b)
}

// UpsertBooks writes many BookSummary rows in a single transaction.
func (db *DB) UpsertBooks(books []kavita.BookSummary) error {
	if len(books) == 0 {
		return nil
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	for _, b := range books {
		if err := upsertBook(tx, b); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetBook returns the row for a chapter, or (nil, nil) if absent.
func (db *DB) GetBook(chapterID int) (*kavita.BookSummary, error) {
	row := db.conn.QueryRow(`SELECT `+BOOK_COLUMNS+` FROM books WHERE chapter_id = ?`, chapterID)
	b, err := scanBook(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListBooks returns every book, newest read first.
func (db *DB) ListBooks() ([]kavita.BookSummary, error) {
	rows, err := db.conn.Query(`SELECT ` + BOOK_COLUMNS + ` FROM books ORDER BY last_read_utc DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []kavita.BookSummary
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// DeleteBook removes a row by chapter ID.
func (db *DB) DeleteBook(chapterID int) error {
	_, err := db.conn.Exec(`DELETE FROM books WHERE chapter_id = ?`, chapterID)
	return err
}

// ---------------------------------------------------------------------------
// internals
// ---------------------------------------------------------------------------

const BOOK_COLUMNS = `chapter_id, series_id, volume_id, library_id, title, series_name,
	authors, isbn, thumbnail, pages, pages_read, progress_pct, last_read_utc,
	summary, release_date, word_count, language, blacklisted`

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func upsertBook(e execer, b kavita.BookSummary) error {
	authors, err := json.Marshal(b.Authors)
	if err != nil {
		return fmt.Errorf("marshal authors: %w", err)
	}

	_, err = e.Exec(`
		INSERT INTO books (
			chapter_id, series_id, volume_id, library_id, title, series_name,
			authors, isbn, thumbnail, pages, pages_read, progress_pct,
			last_read_utc, summary, release_date, word_count, language, updated_at,
			blacklisted
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(chapter_id) DO UPDATE SET
			series_id     = excluded.series_id,
			volume_id     = excluded.volume_id,
			library_id    = excluded.library_id,
			title         = excluded.title,
			series_name   = excluded.series_name,
			authors       = excluded.authors,
			isbn          = excluded.isbn,
			thumbnail     = excluded.thumbnail,
			pages         = excluded.pages,
			pages_read    = excluded.pages_read,
			progress_pct  = excluded.progress_pct,
			last_read_utc = excluded.last_read_utc,
			summary       = excluded.summary,
			release_date  = excluded.release_date,
			word_count    = excluded.word_count,
			language      = excluded.language,
			updated_at    = excluded.updated_at
	`,
		b.ChapterID, b.SeriesID, b.VolumeID, b.LibraryID, b.Title, b.SeriesName,
		string(authors), b.ISBN, b.Thumbnail, b.Pages, b.PagesRead, b.ProgressPct,
		nullableTime(b.LastReadUTC), b.Summary, nullableTime(b.ReleaseDate),
		b.WordCount, b.Language, time.Now().UTC().Format(time.RFC3339),
		b.Blacklisted,
	)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanBook(s scanner) (kavita.BookSummary, error) {
	var (
		b             kavita.BookSummary
		authorsJSON   string
		isbn, thumb   sql.NullString
		summary, lang sql.NullString
		lastRead, rel sql.NullString
	)
	var blacklisted int
	err := s.Scan(
		&b.ChapterID, &b.SeriesID, &b.VolumeID, &b.LibraryID, &b.Title, &b.SeriesName,
		&authorsJSON, &isbn, &thumb, &b.Pages, &b.PagesRead, &b.ProgressPct,
		&lastRead, &summary, &rel, &b.WordCount, &lang, &blacklisted,
	)
	if err != nil {
		return b, err
	}

	if authorsJSON != "" {
		if err := json.Unmarshal([]byte(authorsJSON), &b.Authors); err != nil {
			return b, fmt.Errorf("unmarshal authors: %w", err)
		}
	}
	b.ISBN = isbn.String
	b.Thumbnail = thumb.String
	b.Summary = summary.String
	b.Language = lang.String
	b.LastReadUTC = parseTime(lastRead)
	b.ReleaseDate = parseTime(rel)
	b.Blacklisted = blacklisted != 0
	return b, nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(ns sql.NullString) time.Time {
	if !ns.Valid || ns.String == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return time.Time{}
	}
	return t
}

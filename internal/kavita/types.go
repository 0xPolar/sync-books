// Package kavita provides a Go client for the Kavita REST API.
//
// Struct definitions mirror Kavita's OpenAPI schema (v0.8.9.38). Field names
// use Go conventions; JSON tags preserve the wire format Kavita expects.
//
// Generated against: https://raw.githubusercontent.com/Kareadita/Kavita/develop/openapi.json
package kavita

import "time"

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// MangaFormat is the file format of a chapter/series.
type MangaFormat int

const (
	FormatImage   MangaFormat = 0
	FormatArchive MangaFormat = 1
	FormatUnknown MangaFormat = 2
	FormatEpub    MangaFormat = 3
	FormatPdf     MangaFormat = 4
)

// LibraryType is the kind of content a Kavita library holds.
type LibraryType int

const (
	LibraryManga      LibraryType = 0
	LibraryComic      LibraryType = 1
	LibraryBook       LibraryType = 2
	LibraryImage      LibraryType = 3
	LibraryLightNovel LibraryType = 4
	LibraryComicVine  LibraryType = 5
)

// AgeRating mirrors Kavita's AgeRating enum (a subset; -1 == NotApplicable).
type AgeRating int

// PublicationStatus is where a series is in its publication lifecycle.
type PublicationStatus int

const (
	PubOngoing   PublicationStatus = 0
	PubHiatus    PublicationStatus = 1
	PubCompleted PublicationStatus = 2
	PubCancelled PublicationStatus = 3
	PubEnded     PublicationStatus = 4
)

// PersonRole identifies how a Person is credited on a chapter/series.
type PersonRole int

const (
	RoleOther       PersonRole = 1
	RoleWriter      PersonRole = 3
	RolePenciller   PersonRole = 4
	RoleInker       PersonRole = 5
	RoleColorist    PersonRole = 6
	RoleLetterer    PersonRole = 7
	RoleCoverArtist PersonRole = 8
	RoleEditor      PersonRole = 9
	RolePublisher   PersonRole = 10
	RoleCharacter   PersonRole = 11
	RoleTranslator  PersonRole = 12
	RoleTeam        PersonRole = 13
	RoleLocation    PersonRole = 14
	RoleImprint     PersonRole = 15
)

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

// UserDto is returned from /api/Plugin/authenticate and /api/Account/login.
// The Token field is what you pass as `Authorization: Bearer <Token>`.
type UserDto struct {
	Username       string    `json:"username"`
	Email          string    `json:"email,omitempty"`
	Token          string    `json:"token"`
	RefreshToken   string    `json:"refreshToken"`
	ApiKey         string    `json:"apiKey"`
	Preferences    any       `json:"preferences,omitempty"`
	AgeRestriction any       `json:"ageRestriction,omitempty"`
	KavitaVersion  string    `json:"kavitaVersion,omitempty"`
	Created        time.Time `json:"created,omitempty"`
}

// TokenRequestDto is used for /api/Account/refresh-token.
type TokenRequestDto struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// ---------------------------------------------------------------------------
// Core entities
// ---------------------------------------------------------------------------

// LibraryDto is a Kavita library (a collection of folders the scanner watches).
type LibraryDto struct {
	ID                          int         `json:"id"`
	Name                        string      `json:"name"`
	Type                        LibraryType `json:"type"`
	LastScanned                 time.Time   `json:"lastScanned"`
	CoverImage                  string      `json:"coverImage,omitempty"`
	FolderWatching              bool        `json:"folderWatching"`
	IncludeInDashboard          bool        `json:"includeInDashboard"`
	IncludeInRecommended        bool        `json:"includeInRecommended"`
	ManageCollections           bool        `json:"manageCollections"`
	ManageReadingLists          bool        `json:"manageReadingLists"`
	IncludeInSearch             bool        `json:"includeInSearch"`
	AllowScrobbling             bool        `json:"allowScrobbling"`
	Folders                     []string    `json:"folders,omitempty"`
	CollapseSeriesRelationships bool        `json:"collapseSeriesRelationships"`
	AllowMetadataMatching       bool        `json:"allowMetadataMatching"`
	DefaultLanguage             string      `json:"defaultLanguage,omitempty"`
}

// SeriesDto is one series (a book series, manga, comic run, etc.).
//
// PagesRead / Pages let you compute a per-series read fraction. For
// per-book ISBN/thumbnail/author detail you need to walk into Volumes/Chapters.
type SeriesDto struct {
	ID                  int         `json:"id"`
	Name                string      `json:"name"`
	OriginalName        string      `json:"originalName,omitempty"`
	LocalizedName       string      `json:"localizedName,omitempty"`
	SortName            string      `json:"sortName,omitempty"`
	Pages               int         `json:"pages"`
	CoverImageLocked    bool        `json:"coverImageLocked"`
	LastChapterAdded    time.Time   `json:"lastChapterAdded"`
	LastChapterAddedUtc time.Time   `json:"lastChapterAddedUtc"`
	UserRating          float32     `json:"userRating"`
	HasUserRated        bool        `json:"hasUserRated"`
	TotalReads          int         `json:"totalReads"`
	PagesRead           int         `json:"pagesRead"`
	LatestReadDate      time.Time   `json:"latestReadDate"`
	Format              MangaFormat `json:"format"`
	Created             time.Time   `json:"created"`
	WordCount           int64       `json:"wordCount"`
	LibraryID           int         `json:"libraryId"`
	LibraryName         string      `json:"libraryName,omitempty"`
	MinHoursToRead      int         `json:"minHoursToRead"`
	MaxHoursToRead      int         `json:"maxHoursToRead"`
	AvgHoursToRead      float32     `json:"avgHoursToRead"`
	FolderPath          string      `json:"folderPath,omitempty"`
	CoverImage          string      `json:"coverImage,omitempty"`
	PrimaryColor        string      `json:"primaryColor,omitempty"`
	SecondaryColor      string      `json:"secondaryColor,omitempty"`
	AniListID           int         `json:"aniListId,omitempty"`
	MalID               int64       `json:"malId,omitempty"`
}

// VolumeDto groups Chapters within a Series.
type VolumeDto struct {
	ID              int          `json:"id"`
	MinNumber       float32      `json:"minNumber"`
	MaxNumber       float32      `json:"maxNumber"`
	Name            string       `json:"name,omitempty"`
	Pages           int          `json:"pages"`
	PagesRead       int          `json:"pagesRead"`
	LastModifiedUtc time.Time    `json:"lastModifiedUtc"`
	Created         time.Time    `json:"created"`
	CreatedUtc      time.Time    `json:"createdUtc"`
	SeriesID        int          `json:"seriesId"`
	Chapters        []ChapterDto `json:"chapters,omitempty"`
	CoverImage      string       `json:"coverImage,omitempty"`
}

// ChapterDto is a single chapter (or, for epub/pdf libraries, a single book).
// This is where ISBN, individual cover image, and per-book authors live.
type ChapterDto struct {
	ID                     int               `json:"id"`
	Range                  string            `json:"range,omitempty"`
	MinNumber              float32           `json:"minNumber"`
	MaxNumber              float32           `json:"maxNumber"`
	SortOrder              float32           `json:"sortOrder"`
	Pages                  int               `json:"pages"`
	IsSpecial              bool              `json:"isSpecial"`
	Title                  string            `json:"title,omitempty"`
	Files                  []MangaFileDto    `json:"files,omitempty"`
	PagesRead              int               `json:"pagesRead"`
	TotalReads             int               `json:"totalReads"`
	LastReadingProgressUtc time.Time         `json:"lastReadingProgressUtc"`
	LastReadingProgress    time.Time         `json:"lastReadingProgress"`
	CoverImageLocked       bool              `json:"coverImageLocked"`
	VolumeID               int               `json:"volumeId"`
	CreatedUtc             time.Time         `json:"createdUtc"`
	LastModifiedUtc        time.Time         `json:"lastModifiedUtc"`
	Created                time.Time         `json:"created"`
	ReleaseDate            time.Time         `json:"releaseDate"`
	TitleName              string            `json:"titleName,omitempty"`
	Summary                string            `json:"summary,omitempty"`
	AgeRating              AgeRating         `json:"ageRating"`
	WordCount              int64             `json:"wordCount"`
	VolumeTitle            string            `json:"volumeTitle,omitempty"`
	MinHoursToRead         int               `json:"minHoursToRead"`
	MaxHoursToRead         int               `json:"maxHoursToRead"`
	AvgHoursToRead         float32           `json:"avgHoursToRead"`
	WebLinks               string            `json:"webLinks,omitempty"`
	ISBN                   string            `json:"isbn,omitempty"`
	Writers                []PersonDto       `json:"writers,omitempty"`
	CoverArtists           []PersonDto       `json:"coverArtists,omitempty"`
	Publishers             []PersonDto       `json:"publishers,omitempty"`
	Characters             []PersonDto       `json:"characters,omitempty"`
	Pencillers             []PersonDto       `json:"pencillers,omitempty"`
	Inkers                 []PersonDto       `json:"inkers,omitempty"`
	Colorists              []PersonDto       `json:"colorists,omitempty"`
	Letterers              []PersonDto       `json:"letterers,omitempty"`
	Editors                []PersonDto       `json:"editors,omitempty"`
	Translators            []PersonDto       `json:"translators,omitempty"`
	Genres                 []GenreTagDto     `json:"genres,omitempty"`
	Tags                   []TagDto          `json:"tags,omitempty"`
	PublicationStatus      PublicationStatus `json:"publicationStatus"`
	Language               string            `json:"language,omitempty"`
	Count                  int               `json:"count"`
	TotalCount             int               `json:"totalCount"`
	CoverImage             string            `json:"coverImage,omitempty"`
}

// MangaFileDto describes the on-disk file backing a chapter.
type MangaFileDto struct {
	ID        int         `json:"id"`
	FilePath  string      `json:"filePath,omitempty"`
	Pages     int         `json:"pages"`
	Bytes     int64       `json:"bytes"`
	Format    MangaFormat `json:"format"`
	Created   time.Time   `json:"created"`
	Extension string      `json:"extension,omitempty"`
}

// PersonDto is anyone credited on a work — author, artist, character, etc.
type PersonDto struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	CoverImage  string       `json:"coverImage,omitempty"`
	Aliases     []string     `json:"aliases,omitempty"`
	Description string       `json:"description,omitempty"`
	AsIN        string       `json:"asin,omitempty"`
	AniListID   int          `json:"aniListId,omitempty"`
	MalID       int64        `json:"malId,omitempty"`
	HardcoverID string       `json:"hardcoverId,omitempty"`
	WebLinks    []string     `json:"webLinks,omitempty"`
	Roles       []PersonRole `json:"roles,omitempty"`
}

// GenreTagDto, TagDto are simple {id, title} pairs.
type GenreTagDto struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type TagDto struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ---------------------------------------------------------------------------
// Reading progress
// ---------------------------------------------------------------------------

// ProgressDto is a single read-position record (one user, one chapter).
type ProgressDto struct {
	VolumeID        int       `json:"volumeId"`
	ChapterID       int       `json:"chapterId"`
	PageNum         int       `json:"pageNum"`
	SeriesID        int       `json:"seriesId"`
	LibraryID       int       `json:"libraryId"`
	BookScrollID    string    `json:"bookScrollId,omitempty"`
	LastModifiedUtc time.Time `json:"lastModifiedUtc"`
}

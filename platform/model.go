package platform

import (
	"time"

	"go.uber.org/zap"
)

// PodcastMeta is
type PodcastMeta struct {
	Provider      string // e.g.: 喜马拉雅，荔枝...
	ID            string `storm:"id"`
	Title         string
	Link          string
	Description   string
	Category      []string
	LastBuildDate time.Time
	PubDate       time.Time
	CoverImgURL   string
	IAuthor       string
	ISummary      string

	CdnAudioCover string // lizhifm specific
	Band          string // lizhifm specific
}

// PodcastItem is
type PodcastItem struct {
	Title       string
	PubDate     time.Time
	Description string
	Link        string
	ImageURL    string
	Duration    int
	Src         string
	ID          string `storm:"id"`
	AlbumID     string `storm:"index"`
	AlbumName   string
}

// Podcast is
type Podcast struct {
	meta     PodcastMeta
	items    []PodcastItem
	log      *zap.SugaredLogger
	fetchAll bool
	db       *DB
}

// IPodcastMeta is
type IPodcastMeta interface {
	Meta() PodcastMeta
}

// IPodcastItems is
type IPodcastItems interface {
	Items() []PodcastItem
}

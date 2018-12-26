package platform

import "time"

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
	ID          int
	AlbumID     int
	AlbumName   string
}

// IPodcastMeta is
type IPodcastMeta interface {
	Meta() PodcastMeta
}

// IPodcastItems is
type IPodcastItems interface {
	Items() []PodcastItem
}

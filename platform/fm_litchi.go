package platform

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/levigross/grequests"
	"go.uber.org/zap"
)

// Litchi is 荔枝FM
type Litchi Podcast

type litchiPodcastTrack struct {
	ID               string
	RID              string
	Name             string
	URL              string
	Cover            string
	Duration         int
	CreateTime       int64 `json:"create_time"`
	FixedHighPlayURL string
	FixedLowPlayURL  string
}

type litchiMetaResponse struct {
	CdnAudioCover string
	CdnRadioCover string
	CdnPortrait   string
	Radio         struct {
		Name       string
		Desc       string
		Cover      string
		CreateTime int64
		Band       string
	}
	User struct {
		Portrait string
	}
}

type litchiTrackListResponse struct {
	Total  int
	Audios []litchiPodcastTrack
	P      int
	Size   int
}

// NewLitchi is
func NewLitchi(url, provider string, logger *zap.SugaredLogger, db *DB) *Litchi {
	pid := extractIDFromURL(url)
	return &Litchi{
		meta: PodcastMeta{
			Provider: provider,
			ID:       pid,
			Link:     url,
		},
		log: logger,
		db:  db,
	}
}

var litchiReqOpt = requestOptions(litchiDomain)

// FetchAll is
func (l *Litchi) FetchAll(f bool) *Litchi {
	l.fetchAll = f
	return l
}

func (l *Litchi) fetchMeta() error {
	l.log.Infow("fetching meta information of podcast", "provider", l.meta.Provider, "id", l.meta.ID)

	resp, err := grequests.Get(fmt.Sprintf(litchiPodcastMetaQuery, l.meta.ID), litchiReqOpt)
	if err != nil || resp.StatusCode != 200 {
		l.log.Error(err)
		return err
	}

	var meta litchiMetaResponse
	err = resp.JSON(&meta)
	if err != nil {
		l.log.Error(err)
		return err
	}

	l.meta.Title = meta.Radio.Name
	l.meta.Description = meta.Radio.Desc
	l.meta.Band = meta.Radio.Band
	// TODO where is litchi category
	l.meta.Category = []string{"无"}

	// TODO hard code herer, maybe issue here
	l.meta.CoverImgURL = strings.Replace(fmt.Sprintf("%s%s", meta.CdnPortrait, meta.User.Portrait), ".jpg", "_160x160.jpg", 1)
	l.meta.PubDate = time.Unix(meta.Radio.CreateTime/1000, 0)
	l.meta.LastBuildDate = l.meta.PubDate
	l.meta.CdnAudioCover = meta.CdnAudioCover

	return nil
}

func (l Litchi) getPageRange() (int, error) {
	resp, err := grequests.Get(fmt.Sprintf(litchiPodcastQuery, l.meta.ID, 1), litchiReqOpt)
	if err != nil || resp.StatusCode != 200 {
		l.log.Error(err)
		return 0, err
	}
	var trackList litchiTrackListResponse
	err = resp.JSON(&trackList)
	if err != nil {
		l.log.Error(err)
		return 0, err
	}
	t, s := trackList.Total, trackList.Size

	if t < s {
		return 1, nil
	} else if t%s == 0 {
		return t / s, nil
	}
	return t/s + 1, nil
}

func (l Litchi) fetchTrackDescription(trackID string) (string, error) {
	l.log.Debugw("start fetching track description", "trackID", trackID)
	resp, err := grequests.Get(fmt.Sprintf(litchiTrackInfoQuery, l.meta.Band, trackID), litchiReqOpt)
	if err != nil || resp.StatusCode != 200 {
		l.log.Error(err)
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		l.log.Error(err)
		return "", err
	}
	ret := doc.Find(".desText")
	return ret.Text(), nil
}

func (l *Litchi) fetchTrackList(pageNum int) error {
	count, err := l.getPageRange()
	re := regexp.MustCompile("cdn([0-9]+)")

	if err != nil {
		l.log.Error(err)
		return err
	}

	if !l.fetchAll {
		l.log.Infow("only fetch latests items of podcast", "provider", l.meta.Provider, "id", l.meta.ID)
		count = pageNum
	}
	l.log.Infow("fetch all items of podcast", "provider", l.meta.Provider, "id", l.meta.ID)
	l.log.Debugf("current count value is %d", count)

	for index := pageNum; index <= count; index++ {
		l.log.Debugf("start parsing page %d", index)
		resp, err := grequests.Get(fmt.Sprintf(litchiPodcastQuery, l.meta.ID, index), litchiReqOpt)
		if err != nil || resp.StatusCode != 200 {
			l.log.Error(err)
			continue
		}

		var trackList litchiTrackListResponse
		err = resp.JSON(&trackList)
		if err != nil {
			l.log.Error(err)
			continue
		}
		for _, track := range trackList.Audios {
			desc, err := l.fetchTrackDescription(track.ID)
			l.log.Debugf("get desc %s", desc)
			if err != nil || desc == "" {
				l.log.Error(err)
				desc = track.Name
			}
			item := PodcastItem{
				Title:       track.Name,
				Link:        track.URL,
				ImageURL:    fmt.Sprintf("%s%s", l.meta.CdnAudioCover, track.Cover),
				Duration:    track.Duration,
				Src:         re.ReplaceAllString(track.URL, "cdn"),
				ID:          track.ID,
				AlbumID:     l.meta.ID,
				AlbumName:   l.meta.Title,
				PubDate:     time.Unix(track.CreateTime/1000, 0),
				Description: desc,
			}
			l.log.Debugf("fetched track %s", track.Name)
			l.items = append(l.items, item)
		}
	}
	return nil
}

// Start is
func (l Litchi) Start() error {
	if err := l.fetchMeta(); err != nil {
		return err
	}
	_, err := l.db.FindPodcastMeta(l.meta.ID)
	if err != nil {
		l.log.Warnw("can't find podcast meta info in database", "id", l.meta.ID)
		l.log.Warnw("start a full fetch for podcast", "id", l.meta.ID)
		l.fetchAll = true
	}
	l.log.Infow("found podcast info in database", "id", l.meta.ID)
	l.log.Infow("only fetch latest for podcast", "id", l.meta.ID)
	l.log.Infof("fetch all switch now is %v", l.fetchAll)

	if err := l.fetchTrackList(1); err != nil {
		return err
	}

	l.log.Info("save fetched data into database")
	if err := l.db.SaveMetaData(l); err != nil {
		return err
	}
	if err := l.db.SaveItems(l); err != nil {
		return err
	}
	l.log.Info("start making rss feed file")
	ProduceRSSFeed(l.meta.ID, l.db, l.log)

	return nil
}

// Implement IPodcastMeta and IPodcastItems interface

// Meta is
func (l Litchi) Meta() PodcastMeta {
	return l.meta
}

// Items is
func (l Litchi) Items() []PodcastItem {
	return l.items
}

package platform

import (
	"fmt"
	"strconv"
	"time"

	"github.com/levigross/grequests"
	"go.uber.org/zap"
)

// Himalaya is
type Himalaya Podcast

type himalayaPodcastTrack struct {
	TrackID        int `storm:"id"`
	TrackName      string
	TrackURL       string
	TrackCoverPath string
	Duration       int
	Src            string
	AlbumName      string
	AlbumID        int `storm:"index"`
}

type (
	himalayaMetaResponse struct {
		Ret  int
		Data struct {
			AlbumID  int
			MainInfo struct {
				Cover      string
				AlbumTitle string
				Crumbs     struct {
					CategoryPinyin  string
					SubcategoryCode string
				}
				UpdateDate      string
				RichIntro       string
				DetailRichIntro string
			}
		}
	}

	himalayaTrackListResponse struct {
		Ret  int
		Data struct {
			TracksAudioPlay []himalayaPodcastTrack
			HasMore         bool
		}
	}

	himalayaTrackResponse struct {
		Ret  int
		Data struct {
			TrackInfo struct {
				RichIntro  string
				Draft      string
				LastUpdate string
			}
		}
	}
)

// NewHimalaya is
func NewHimalaya(url,
	provider string,
	logger *zap.SugaredLogger,
	db *DB) *Himalaya {
	pid := extractIDFromURL(url)
	return &Himalaya{
		meta: PodcastMeta{
			Provider: provider,
			ID:       pid,
			Link:     url,
		},
		log: logger,
		db:  db,
	}
}

// FetchAll if fetch all items or only latest
func (h *Himalaya) FetchAll(all bool) *Himalaya {
	h.fetchAll = all
	return h
}

func (h *Himalaya) fetchMeta() error {
	h.log.Infow("fetching meta information of podcast", "provider", h.meta.Provider, "id", h.meta.ID)

	reqOpt := requestOptions(himalayaDomain)
	resp, err := grequests.Get(fmt.Sprintf(himalayaPodcastMetaQuery, h.meta.ID), reqOpt)
	if err != nil || resp.StatusCode != 200 {
		h.log.Error(err)
		return err
	}

	var meta himalayaMetaResponse
	err = resp.JSON(&meta)
	if err != nil {
		h.log.Error(err)
		return err
	}

	h.meta.Title = meta.Data.MainInfo.AlbumTitle
	h.meta.Description = meta.Data.MainInfo.RichIntro
	h.meta.Category = append(h.meta.Category, meta.Data.MainInfo.Crumbs.SubcategoryCode)
	date, err := time.Parse(himalayaTimeLayoutShort, meta.Data.MainInfo.UpdateDate)
	if err != nil {
		h.log.Error(err)
	}
	h.meta.LastBuildDate = date
	// TODO get real pubDate, now just minus 2 years
	h.meta.PubDate = date.AddDate(-2, 0, 0)
	h.meta.CoverImgURL = fmt.Sprintf("http:%s", meta.Data.MainInfo.Cover)
	h.meta.ISummary = meta.Data.MainInfo.DetailRichIntro
	return nil
}

func (h Himalaya) fetchTrackMeta(trackID int) (string, time.Time, error) {
	reqOpt := requestOptions(himalayaDomain)
	resp, err := grequests.Get(fmt.Sprintf(himalayaItemQuery, trackID), reqOpt)
	if err != nil || resp.StatusCode != 200 {
		h.log.Error(err)
		return "", time.Now(), err
	}
	var track himalayaTrackResponse
	err = resp.JSON(&track)
	if err != nil {
		return "", time.Now(), err
	}
	pubDate, err := time.Parse(himalayaTimeLayout, track.Data.TrackInfo.LastUpdate)
	if err != nil {
		h.log.Error(err)
		pubDate = time.Now()
	}
	if track.Data.TrackInfo.Draft == "" && track.Data.TrackInfo.RichIntro == "" {
		return "no description", pubDate, nil
	} else if track.Data.TrackInfo.RichIntro == "" {
		return track.Data.TrackInfo.Draft, pubDate, nil
	}
	return track.Data.TrackInfo.RichIntro, pubDate, nil
}

func (h *Himalaya) fetchTrackList(pageNum int) error {
	h.log.Debugf("fetching tracklist from page %d", pageNum)

	reqOpt := requestOptions(himalayaDomain)
	var trackList himalayaTrackListResponse

	resp, err := grequests.Get(fmt.Sprintf(himalayaPodcastQuery, h.meta.ID, pageNum), reqOpt)
	if err != nil || resp.StatusCode != 200 {
		h.log.Error(err)
		return err
	}
	err = resp.JSON(&trackList)
	if err != nil {
		h.log.Error(err)
		return err
	}
	for _, track := range trackList.Data.TracksAudioPlay {
		desc, pubDate, err := h.fetchTrackMeta(track.TrackID)
		if err != nil {
			h.log.Error(err)
		}
		item := PodcastItem{
			Title:       track.TrackName,
			Link:        fmt.Sprintf("https://www.ximalaya.com%s", track.TrackURL),
			ImageURL:    fmt.Sprintf("http:%s", track.TrackCoverPath),
			Duration:    track.Duration,
			Src:         track.Src,
			ID:          strconv.Itoa(track.TrackID),
			AlbumID:     strconv.Itoa(track.AlbumID),
			AlbumName:   track.AlbumName,
			PubDate:     pubDate,
			Description: desc,
		}
		h.log.Debugf("fetched track %s", track.TrackName)
		h.items = append(h.items, item)
	}

	if !h.fetchAll {
		return nil
	}

	if trackList.Data.HasMore {
		return h.fetchTrackList(pageNum + 1)
	}

	return nil
}

// Start is
func (h Himalaya) Start() error {
	if err := h.fetchMeta(); err != nil {
		return err
	}
	_, err := h.db.FindPodcastMeta(h.meta.ID)
	if err != nil {
		h.log.Warnw("can't find podcast meta info in database", "id", h.meta.ID)
		h.log.Warnw("start a full fetch for podcast", "id", h.meta.ID)
		h.fetchAll = true
	}
	h.log.Infow("found podcast info in database", "id", h.meta.ID)
	h.log.Infow("only fetch latest for podcast", "id", h.meta.ID)
	h.log.Infof("fetch all switch now is %v", h.fetchAll)
	if err := h.fetchTrackList(1); err != nil {
		return err
	}
	h.log.Info("save fetched data into database")
	if err := h.db.SaveMetaData(h); err != nil {
		return err
	}
	if err := h.db.SaveItems(h); err != nil {
		return err
	}
	h.log.Info("start making rss feed file")
	ProduceRSSFeed(h.meta.ID, h.db, h.log)
	return nil
}

// Implement IPodcastMeta and IPodcastItems interface

// Meta is
func (h Himalaya) Meta() PodcastMeta {
	return h.meta
}

// Items is
func (h Himalaya) Items() []PodcastItem {
	return h.items
}

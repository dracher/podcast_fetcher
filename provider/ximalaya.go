package provider

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"

	"github.com/eduncan911/podcast"
	"github.com/levigross/grequests"
)

var ximalayaReqOptions = requestOptions("www.ximalaya.com")

type (
	// XimalayaTrack is
	XimalayaTrack struct {
		TrackID        int `storm:"id"`
		TrackName      string
		TrackURL       string
		TrackCoverPath string
		Duration       int
		Src            string
		AlbumName      string
		AlbumID        int `storm:"index"`
	}
	// XimalayaAlbum is
	XimalayaAlbum struct {
		ID            int
		Name          string
		Desc          string
		Link          string
		CoverImgURL   string
		PubDate       time.Time
		LastBuildDate time.Time
		trackList     []XimalayaTrack
		DB            *storm.DB
	}

	// XimalayaAlbumResponse is
	XimalayaAlbumResponse struct {
		Ret  int
		Data struct {
			TracksAudioPlay []XimalayaTrack
			HasMore         bool
		}
	}

	// XimalayaTrackResponse is
	XimalayaTrackResponse struct {
		Ret  int
		Data struct {
			TrackInfo struct {
				RichIntro  string
				Draft      string
				LastUpdate string
			}
		}
	}

	// XimalayaAlbumMetaResponse is
	XimalayaAlbumMetaResponse struct {
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
)

// NewXimalayaAlbum is
func NewXimalayaAlbum(albumURL string, db *storm.DB) *XimalayaAlbum {
	aid := GetAlbumURLastPart(albumURL)

	ximalaya := XimalayaAlbum{DB: db}

	ximalaya.fetchXimalayaMeta(aid)
	ximalaya.Link = albumURL

	return &ximalaya
}

func (x *XimalayaAlbum) fetchXimalayaMeta(albumID string) {
	ret, err := grequests.Get(fmt.Sprintf(ximalayaAlbumMetaQuery, albumID), ximalayaReqOptions)
	if err != nil {
		log.Fatal(err)
	}
	var meta XimalayaAlbumMetaResponse
	ret.JSON(&meta)
	x.ID = meta.Data.AlbumID
	x.Name = meta.Data.MainInfo.AlbumTitle
	x.Desc = meta.Data.MainInfo.RichIntro
	x.CoverImgURL = fmt.Sprintf("http:%s", meta.Data.MainInfo.Cover)
	date, _ := time.Parse(ximalayaTimeLayoutShort, meta.Data.MainInfo.UpdateDate)
	x.LastBuildDate = date
	x.PubDate = date.AddDate(-2, 0, 0)

}

func (x XimalayaAlbum) fetchTrackByID(trackID int) ([3]string, error) {
	ret, err := grequests.Get(fmt.Sprintf(ximalayaTrackQuery, trackID), ximalayaReqOptions)
	if err != nil {
		log.Error(err)
		return [3]string{}, err
	}

	var resp XimalayaTrackResponse
	err = ret.JSON(&resp)
	if err != nil || resp.Ret != 200 {
		log.Error(err)
		return [3]string{}, err
	}
	return [3]string{resp.Data.TrackInfo.RichIntro, resp.Data.TrackInfo.LastUpdate, resp.Data.TrackInfo.Draft}, nil
}

func (x *XimalayaAlbum) fetchAlbumByID(pageNum int, fetchAll bool) error {
	log.Debug(pageNum)
	var resp XimalayaAlbumResponse

	ret, err := grequests.Get(fmt.Sprintf(ximalayaAlbumQuery, x.ID, pageNum), ximalayaReqOptions)
	if err != nil {
		log.Error(err)
		return err
	}

	err = ret.JSON(&resp)
	if err != nil || resp.Ret != 200 {
		log.Error(err)
		return err
	}

	x.trackList = append(x.trackList, resp.Data.TracksAudioPlay...)

	if !fetchAll {
		return nil
	}

	if resp.Data.HasMore {
		return x.fetchAlbumByID(pageNum+1, false)
	}
	return nil
}

// ProduceRSSFeed is
func (x XimalayaAlbum) ProduceRSSFeed(pageNum int, fetchAll bool) error {
	p := podcast.New(x.Name, x.Link, x.Desc, &x.PubDate, &x.LastBuildDate)
	p.AddImage(x.CoverImgURL)

	err := x.fetchAlbumByID(pageNum, fetchAll)
	if err != nil {
		return err
	}

	for _, track := range x.trackList {
		track.AlbumName = x.Name
		track.AlbumID = x.ID
		err := x.DB.Save(&track)
		if err != nil {
			log.Error(err)
		}
	}

	// for _, track := range x.trackList {
	// 	log.Debugf("processing %s", track.TrackName)
	// 	il, err := x.fetchTrackByID(track.TrackID)
	// 	if err != nil {
	// 		log.Error(err)
	// 		continue
	// 	}
	// 	pubDate, _ := time.Parse(ximalayaTimeLayout, il[1])

	// 	var desc string
	// 	if il[0] == "" && il[2] == "" {
	// 		desc = "no description provider"
	// 	} else if il[0] == "" {
	// 		desc = il[2]
	// 	} else {
	// 		desc = il[0]
	// 	}

	// 	item := podcast.Item{
	// 		Title:       track.TrackName,
	// 		PubDate:     &pubDate,
	// 		Description: desc,
	// 		Link:        fmt.Sprintf("https://www.ximalaya.com%s", track.TrackURL),
	// 	}
	// 	item.AddImage(fmt.Sprintf("http:%s", track.TrackCoverPath))
	// 	item.AddDuration(int64(track.Duration))
	// 	item.AddEnclosure(track.Src, podcast.M4A, 0)

	// 	_, err = p.AddItem(item)
	// 	if err != nil {
	// 		log.Error(err, track)
	// 		continue
	// 	}
	// }

	// fp, _ := os.OpenFile(fmt.Sprintf("ximalaya_%d.xml", x.ID), os.O_RDWR|os.O_CREATE, 0755)
	// defer fp.Close()

	// if err := p.Encode(fp); err != nil {
	// 	fmt.Println("error writing to stdout:", err.Error())
	// }
	return nil
}

package provider

import (
	"fmt"
	"os"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/levigross/grequests"
	log "github.com/sirupsen/logrus"
)

const (
	lizhiUserAudioQuery = "http://www.lizhi.fm/api/user/audios/%s/%d"
	lizhiUserInfoQuery  = "http://www.lizhi.fm/api/user/info/%s"
	lizhiSingleURL      = "http://www.lizhi.fm/%s/%s"
)

var lizhiReqOptions = grequests.RequestOptions{
	Headers: map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent":                "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0",
		"Host":                      "www.lizhi.fm",
		"Upgrade-Insecure-Requests": "1",
	},
}

type (
	// LizhiUser is
	LizhiUser struct {
		ID            string
		Name          string
		Desc          string
		Link          string
		CoverImgURL   string
		PubDate       time.Time
		LastBuildDate time.Time
		CdnAudioCover string
		BandID        string
		trackList     []struct {
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
	}

	// LizhiUserResponse is
	LizhiUserResponse struct {
		Total  int
		Audios []struct {
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
		P    int
		Size int
	}

	// LizhiUserMetaInfoResponse is
	LizhiUserMetaInfoResponse struct {
		CdnAudioCover string
		CdnRadioCover string
		Radio         struct {
			Name       string
			Desc       string
			Cover      string
			CreateTime int64
			Band       string
		}
		User struct {
			Name  string
			Email string
		}
	}
)

// NewLizhiUser is
func NewLizhiUser(albumURL string) *LizhiUser {
	aid := GetAlbumURLastPart(albumURL)
	lu := LizhiUser{ID: aid, Link: albumURL}

	err := lu.fetchLizhiMeta()
	if err != nil {
		log.Fatal(err)
	}
	err = lu.fetchUserByID()
	if err != nil {
		log.Fatal(err)
	}
	return &lu
}

func (l *LizhiUser) fetchUserByID() error {
	ret, err := grequests.Get(fmt.Sprintf(lizhiUserAudioQuery, l.ID, 1), &lizhiReqOptions)
	if err != nil {
		log.Error(err)
		return err
	}
	var resp LizhiUserResponse
	err = ret.JSON(&resp)
	if err != nil {
		log.Error(err)
		return err
	}

	var count int
	if resp.Total < resp.Size {
		count = 1
	} else if resp.Total%resp.Size == 0 {
		count = resp.Total / resp.Size
	} else {
		count = resp.Total/resp.Size + 1
	}

	for index := 1; index <= count; index++ {
		ret, err := grequests.Get(fmt.Sprintf(lizhiUserAudioQuery, l.ID, index), &lizhiReqOptions)
		if err != nil {
			log.Error(err)
			continue
		}
		var resp LizhiUserResponse
		err = ret.JSON(&resp)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("parsing page %d", index)
		l.trackList = append(l.trackList, resp.Audios...)
		time.Sleep(time.Millisecond * 500)
	}
	return nil
	// data, _ := json.Marshal(l.trackList)
	// ioutil.WriteFile("lz.json", data, 0755)
}

func (l *LizhiUser) fetchLizhiMeta() error {
	ret, err := grequests.Get(fmt.Sprintf(lizhiUserInfoQuery, l.ID), &lizhiReqOptions)
	if err != nil {
		log.Error(err)
		return err
	}

	var metaInfo LizhiUserMetaInfoResponse
	err = ret.JSON(&metaInfo)
	if err != nil {
		log.Error(err)
		return err
	}

	l.Name = metaInfo.Radio.Name
	l.Desc = metaInfo.Radio.Desc
	l.CdnAudioCover = metaInfo.CdnAudioCover
	l.BandID = metaInfo.Radio.Band
	l.CoverImgURL = fmt.Sprintf("%s%s", metaInfo.CdnRadioCover, metaInfo.Radio.Cover)
	l.PubDate = time.Unix(metaInfo.Radio.CreateTime/1000, 0)
	l.LastBuildDate = time.Now()

	return nil
}

// ProduceRSSFeed is
func (l LizhiUser) ProduceRSSFeed() error {
	p := podcast.New(l.Name, l.Link, l.Desc, &l.PubDate, &l.LastBuildDate)
	p.AddImage(l.CoverImgURL)

	for _, track := range l.trackList {
		pubDate := time.Unix(track.CreateTime/1000, 0)
		item := podcast.Item{
			Title:       track.Name,
			PubDate:     &pubDate,
			Description: track.Name,
			Link:        fmt.Sprintf(lizhiSingleURL, l.BandID, track.ID),
		}
		item.AddImage(fmt.Sprintf("%s%s", l.CdnAudioCover, track.Cover))
		item.AddDuration(int64(track.Duration))
		item.AddEnclosure(track.URL, podcast.MP3, 0)

		_, err := p.AddItem(item)
		if err != nil {
			log.Error(err, track)
			continue
		}
	}

	fp, _ := os.OpenFile(fmt.Sprintf("lizhi_%s.xml", l.ID), os.O_RDWR|os.O_CREATE, 0755)
	defer fp.Close()

	if err := p.Encode(fp); err != nil {
		fmt.Println("error writing to stdout:", err.Error())
	}

	return nil
}

package provider

import (
	"errors"
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/levigross/grequests"
	log "github.com/sirupsen/logrus"
)

var lizhiReqOptions = requestOptions("ww.lizhi.fm")

type (
	// LizhiTrack is
	LizhiTrack struct {
		ID               string `storm:"id"`
		RID              string
		Name             string
		URL              string
		Cover            string
		Duration         int
		CreateTime       int64 `json:"create_time"`
		FixedHighPlayURL string
		FixedLowPlayURL  string
		AlbumName        string
		AlbumID          string `storm:"index"`
	}

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
		trackList     []LizhiTrack
		DB            *storm.DB
	}

	// LizhiUserResponse is
	LizhiUserResponse struct {
		Total  int
		Audios []LizhiTrack
		P      int
		Size   int
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
	}
)

// NewLizhiUser is
func NewLizhiUser(albumURL string, db *storm.DB) *LizhiUser {
	aid := GetAlbumURLastPart(albumURL)
	lu := LizhiUser{ID: aid, Link: albumURL, DB: db}

	err := lu.fetchLizhiMeta()
	if err != nil {
		log.Fatal(err, 79)
	}
	return &lu
}

func (l *LizhiUser) fetchUserByID(pageNum int, fetchAll bool) error {
	ret, err := grequests.Get(fmt.Sprintf(lizhiUserAudioQuery, l.ID, 1), lizhiReqOptions)
	if err != nil {
		log.Error(err, 87)
		return err
	}
	var resp LizhiUserResponse
	err = ret.JSON(&resp)
	if err != nil {
		log.Error(err, 93)
		return err
	}

	if !fetchAll {
		log.Info("only fetch latest items")
		return nil
	}

	var count int
	if resp.Total < resp.Size {
		count = 1
	} else if resp.Total%resp.Size == 0 {
		count = resp.Total / resp.Size
	} else {
		count = resp.Total/resp.Size + 1
	}

	for index := pageNum; index <= count; index++ {
		ret, err := grequests.Get(fmt.Sprintf(lizhiUserAudioQuery, l.ID, index), lizhiReqOptions)
		if err != nil {
			log.Error(err, 114)
			continue
		}
		var resp LizhiUserResponse
		err = ret.JSON(&resp)
		if err != nil {
			log.Error(err, 120)
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
	ret, err := grequests.Get(fmt.Sprintf(lizhiUserInfoQuery, l.ID), lizhiReqOptions)
	if err != nil {
		log.Error(err, 135)
		return err
	}

	if ret.StatusCode != 200 {
		return errors.New("page not found")
	}

	var metaInfo LizhiUserMetaInfoResponse

	err = ret.JSON(&metaInfo)
	if err != nil {
		log.Error(err, 142)
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
func (l LizhiUser) ProduceRSSFeed(pageNum int, fetchAll bool) error {
	var res LizhiTrack

	all := checkAlbumAlreadyFetched(res, l.DB, fetchAll)
	log.Info(all)
	// p := podcast.New(l.Name, l.Link, l.Desc, &l.PubDate, &l.LastBuildDate)
	// p.AddImage(l.CoverImgURL)

	// err := l.fetchUserByID(pageNum, fetchAll)
	// if err != nil {
	// 	log.Fatal(err, 164)
	// }

	// for _, track := range l.trackList {
	// 	pubDate := time.Unix(track.CreateTime/1000, 0)
	// 	item := podcast.Item{
	// 		Title:       track.Name,
	// 		PubDate:     &pubDate,
	// 		Description: track.Name,
	// 		Link:        fmt.Sprintf(lizhiSingleURL, l.BandID, track.ID),
	// 	}
	// 	item.AddImage(fmt.Sprintf("%s%s", l.CdnAudioCover, track.Cover))
	// 	item.AddDuration(int64(track.Duration))
	// 	item.AddEnclosure(track.URL, podcast.MP3, 0)

	// 	_, err := p.AddItem(item)
	// 	if err != nil {
	// 		log.Error(err, track)
	// 		continue
	// 	}
	// }

	// fp, _ := os.OpenFile(fmt.Sprintf("lizhi_%s.xml", l.ID), os.O_RDWR|os.O_CREATE, 0755)
	// defer fp.Close()

	// if err := p.Encode(fp); err != nil {
	// 	fmt.Println("error writing to stdout:", err.Error())
	// }

	return nil
}

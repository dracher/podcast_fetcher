package platform

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduncan911/podcast"
	"go.uber.org/zap"
)

func extractIDFromURL(url string) string {
	url = strings.TrimRight(url, "/")
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func getMediaType(url string, log *zap.SugaredLogger) podcast.EnclosureType {
	if strings.HasSuffix(url, "m4a") {
		return podcast.M4A
	} else if strings.HasSuffix(url, "mp3") {
		return podcast.MP3
	} else if strings.HasSuffix(url, "m4v") {
		return podcast.M4V
	} else if strings.HasSuffix(url, "mp4") {
		return podcast.MP4
	} else if strings.HasSuffix(url, "mov") {
		return podcast.MOV
	}
	log.Errorw("url don't match any known media type, return default m4a", "url", url)
	return podcast.M4A
}

// ProduceRSSFeed is
func ProduceRSSFeed(pid string, db *DB, log *zap.SugaredLogger) {
	meta, _ := db.FindPodcastMeta(pid)
	items, _ := db.FindPodcastItems(pid)

	pd := podcast.New(
		meta.Title,
		meta.Link,
		meta.Description,
		&meta.PubDate,
		&meta.LastBuildDate,
	)
	if len(meta.Category) != 0 {
		pd.AddCategory(meta.Category[0], []string{})
	}
	pd.AddImage(meta.CoverImgURL)

	for _, item := range items {
		i := podcast.Item{
			Title:       item.Title,
			PubDate:     &item.PubDate,
			Description: item.Description,
			Link:        item.Link,
		}
		i.AddImage(item.ImageURL)
		i.AddDuration(int64(item.Duration))
		i.AddEnclosure(item.Src, getMediaType(item.Src, log), 0)

		_, err := pd.AddItem(i)
		if err != nil {
			log.Error(err)
		}
	}

	fp, _ := os.OpenFile(fmt.Sprintf("%s.xml", meta.ID), os.O_RDWR|os.O_CREATE, 0755)
	defer fp.Close()
	if err := pd.Encode(fp); err != nil {
		fmt.Println("error writing to stdout:", err.Error())
	}
}

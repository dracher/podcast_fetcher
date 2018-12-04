package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/dracher/podcast_fetcher/provider"
)

var albumURL string
var providerType string

func init() {
	log.SetLevel(log.DebugLevel)
	flag.StringVar(&albumURL, "url", "", "album url, e.g.: https://www.ximalaya.com/yingshi/213124/")
	flag.StringVar(&providerType, "pt", "ximalaya", "provider type, currently only support '喜马拉雅'")
}

func main() {
	flag.Parse()

	if albumURL == "" {
		flag.Usage()
		os.Exit(0)
	}

	switch providerType {
	case "ximalaya":
		ximalaya := provider.NewXimalayaAlbum(albumURL)
		ximalaya.ProduceRSSFeed(1, false)
	default:
		log.Fatal("currently only support 'ximalaya'")
	}

}

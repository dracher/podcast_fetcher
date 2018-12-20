package main

import (
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/dracher/podcast_fetcher/provider"
)

var (
	errURLEmpty = errors.New("url can't be empty")
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func parseURL(uri string) error {
	_, err := url.ParseRequestURI(uri)
	return err
}

func main() {
	db, err := storm.Open("podcasts.db")
	defer db.Close()

	app := cli.NewApp()
	app.Name = "podcast_fetcher"
	app.Version = "0.1.0"
	app.Compiled = time.Now()
	app.Author = "dracher"
	app.Email = "dracher@gmail.com"

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "喜马拉雅",
			ShortName: "xi",
			Usage:     "fetch album from 喜马拉雅",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "url",
					Usage: "e.g.: https://www.ximalaya.com/yingshi/213124/",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "wether fetcher all items, default only fetch latest",
				},
			},
			Before: func(c *cli.Context) error {
				return parseURL(c.String("url"))
			},
			Action: func(c *cli.Context) error {
				log.Info(c.Bool("all"))
				xi := provider.NewXimalayaAlbum(c.String("url"), db)
				return xi.ProduceRSSFeed(1, c.Bool("all"))
			},
		},
		cli.Command{
			Name:      "荔枝FM",
			ShortName: "lz",
			Usage:     "fetch album from 荔枝FM",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "url",
					Usage: "e.g.: http://www.lizhi.fm/user/2554978980702743084",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "whether fetch all items, default only fetch latest",
				},
			},
			Before: func(c *cli.Context) error {
				return parseURL(c.String("url"))
			},
			Action: func(c *cli.Context) error {
				log.Info(c.Bool("all"))
				lz := provider.NewLizhiUser(c.String("url"), db)
				return lz.ProduceRSSFeed(1, c.Bool("all"))
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

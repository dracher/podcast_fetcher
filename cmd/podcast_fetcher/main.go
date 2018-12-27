package main

import (
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/asdine/storm"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/dracher/podcast_fetcher/platform"
)

var (
	errURLEmpty = errors.New("url can't be empty")
	logger      *zap.SugaredLogger
)

func init() {
	log, _ := zap.NewProduction()
	// log, _ := zap.NewDevelopment()
	defer log.Sync()
	logger = log.Sugar()
}

func parseURL(uri string) error {
	_, err := url.ParseRequestURI(uri)
	return err
}

func main() {
	db, err := storm.Open("podcasts.db")
	defer db.Close()
	conn := platform.NewDB(db, logger)

	app := cli.NewApp()
	app.Name = "podcast_fetcher"
	app.Usage = "convert china popular podcaster platform to general podcast format"
	app.Version = "0.2.0"
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
				return platform.NewHimalaya(c.String("url"), c.Command.Name, logger, conn).Start()
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
				return platform.NewLitchi(c.String("url"), c.Command.Name, logger, conn).Start()
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

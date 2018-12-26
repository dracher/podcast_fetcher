package platform

import "github.com/levigross/grequests"

const (
	litchiUserAudioQuery = "http://www.lizhi.fm/api/user/audios/%s/%d"
	litchiUserInfoQuery  = "http://www.lizhi.fm/api/user/info/%s"
	litchiSingleURL      = "http://www.lizhi.fm/%s/%s"
	litchiDomain         = "ww.lizhi.fm"

	himalayaPodcastMetaQuery = "https://www.ximalaya.com/revision/album?albumId=%s"
	himalayaPodcastQuery     = "https://www.ximalaya.com/revision/play/album?albumId=%s&pageNum=%d"
	himalayaItemQuery        = "https://www.ximalaya.com/revision/track/trackPageInfo?trackId=%d"
	himalayaTimeLayout       = "2006-01-02 15:04:05"
	himalayaTimeLayoutShort  = "2006-01-02"
	himalayaDomain           = "www.ximalaya.com"
)

func requestOptions(hostDomain string) *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"User-Agent":                "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0",
			"Host":                      hostDomain,
			"Upgrade-Insecure-Requests": "1",
		},
	}
}

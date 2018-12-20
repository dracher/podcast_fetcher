package provider

import "github.com/levigross/grequests"

const (
	lizhiUserAudioQuery = "http://www.lizhi.fm/api/user/audios/%s/%d"
	lizhiUserInfoQuery  = "http://www.lizhi.fm/api/user/info/%s"
	lizhiSingleURL      = "http://www.lizhi.fm/%s/%s"

	ximalayaAlbumMetaQuery  = "https://www.ximalaya.com/revision/album?albumId=%s"
	ximalayaAlbumQuery      = "https://www.ximalaya.com/revision/play/album?albumId=%d&pageNum=%d"
	ximalayaTrackQuery      = "https://www.ximalaya.com/revision/track/trackPageInfo?trackId=%d"
	ximalayaTimeLayout      = "2006-01-02 15:04:05"
	ximalayaTimeLayoutShort = "2006-01-02"
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

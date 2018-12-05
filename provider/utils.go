package provider

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetAlbumURLastPart is
func GetAlbumURLastPart(albumURL string) string {
	albumURL = strings.TrimRight(albumURL, "/")
	tmpURL := strings.Split(albumURL, "/")
	log.Debugf("trimed url: %s", tmpURL)
	aid := tmpURL[len(tmpURL)-1]
	log.Debugf("album id: %s", aid)
	return aid
}

package platform

import (
	"github.com/asdine/storm"
	"go.uber.org/zap"
)

// DB is
type DB struct {
	db  *storm.DB
	log *zap.SugaredLogger
}

// NewDB is
func NewDB(db *storm.DB, log *zap.SugaredLogger) *DB {
	return &DB{
		db:  db,
		log: log,
	}
}

// SaveMetaData is
func (d DB) SaveMetaData(meta IPodcastMeta) error {
	data := meta.Meta()
	err := d.db.Save(&data)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

// SaveItems is
func (d DB) SaveItems(data IPodcastItems) error {
	for _, item := range data.Items() {
		err := d.db.Save(&item)
		if err != nil {
			d.log.Error(err)
			return err
		}
	}
	return nil
}

// FindPodcastMeta is
func (d DB) FindPodcastMeta(pid string) (PodcastMeta, error) {
	var meta PodcastMeta
	if err := d.db.One("ID", pid, &meta); err != nil {
		d.log.Error(err)
		return meta, err
	}
	return meta, nil
}

// FindPodcastItems is
func (d DB) FindPodcastItems(pid string) (items []PodcastItem, err error) {
	d.db.Find("AlbumID", pid, &items)
	return
}

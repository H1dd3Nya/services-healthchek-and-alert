package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"

	"go.etcd.io/bbolt"
)

type BoltStorage struct {
	db *bbolt.DB
}

var sitesBucket = []byte("sites")
var historyBucket = []byte("history")

func NewBoltStorage(path string) (*BoltStorage, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// Ensure bucket exists
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(sitesBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &BoltStorage{db: db}, nil
}

func (s *BoltStorage) AddSite(site Site) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(sitesBucket)
		data, err := json.Marshal(site)
		if err != nil {
			return err
		}
		return b.Put([]byte(site.ID), data)
	})
	if err != nil {
		log.Printf("[ERROR] Bolt AddSite: %v", err)
	}
	return err
}

func (s *BoltStorage) GetSite(id string) (*Site, error) {
	var site Site
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(sitesBucket)
		v := b.Get([]byte(id))
		if v == nil {
			return errors.New("site not found")
		}
		return json.Unmarshal(v, &site)
	})
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *BoltStorage) ListSites() ([]Site, error) {
	sites := []Site{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(sitesBucket)
		return b.ForEach(func(k, v []byte) error {
			var site Site
			if err := json.Unmarshal(v, &site); err != nil {
				return err
			}
			sites = append(sites, site)
			return nil
		})
	})
	return sites, err
}

func (s *BoltStorage) DeleteSite(id string) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(sitesBucket)
		return b.Delete([]byte(id))
	})
	if err != nil {
		log.Printf("[ERROR] Bolt DeleteSite: %v", err)
	}
	return err
}

func (s *BoltStorage) AddCheckHistory(history SiteCheckHistory) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(historyBucket)
		if err != nil {
			return err
		}
		key := []byte(history.SiteID + ":" + string(rune(history.CheckedAt)))
		data, err := json.Marshal(history)
		if err != nil {
			return err
		}
		return b.Put(key, data)
	})
	if err != nil {
		log.Printf("[ERROR] Bolt AddCheckHistory: %v", err)
	}
	return err
}

func (s *BoltStorage) ListCheckHistory(siteID string, limit int) ([]SiteCheckHistory, error) {
	history := []SiteCheckHistory{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(historyBucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		prefix := []byte(siteID + ":")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var h SiteCheckHistory
			if err := json.Unmarshal(v, &h); err != nil {
				continue
			}
			history = append(history, h)
			if limit > 0 && len(history) >= limit {
				break
			}
		}
		return nil
	})
	return history, err
}

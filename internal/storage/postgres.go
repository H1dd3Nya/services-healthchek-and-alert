package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Миграции: создать таблицы, если их нет
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		check_type TEXT NOT NULL,
		interval_seconds INT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS check_history (
		id SERIAL PRIMARY KEY,
		site_id TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
		checked_at BIGINT NOT NULL,
		duration_ms BIGINT NOT NULL,
		http_code INT,
		success BOOL NOT NULL,
		error TEXT
	);
	`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) AddSite(site Site) error {
	_, err := s.db.Exec(`INSERT INTO sites (id, name, url, check_type, interval_seconds) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id) DO UPDATE SET name=$2, url=$3, check_type=$4, interval_seconds=$5`,
		site.ID, site.Name, site.URL, string(site.CheckType), site.IntervalSeconds)
	if err != nil {
		log.Printf("[ERROR] Postgres AddSite: %v", err)
	}
	return err
}

func (s *PostgresStorage) GetSite(id string) (*Site, error) {
	row := s.db.QueryRow(`SELECT id, name, url, check_type, interval_seconds FROM sites WHERE id=$1`, id)
	site := Site{}
	var checkType string
	if err := row.Scan(&site.ID, &site.Name, &site.URL, &checkType, &site.IntervalSeconds); err != nil {
		return nil, err
	}
	site.CheckType = CheckType(checkType)
	return &site, nil
}

func (s *PostgresStorage) ListSites() ([]Site, error) {
	rows, err := s.db.Query(`SELECT id, name, url, check_type, interval_seconds FROM sites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sites := []Site{}
	for rows.Next() {
		site := Site{}
		var checkType string
		if err := rows.Scan(&site.ID, &site.Name, &site.URL, &checkType, &site.IntervalSeconds); err != nil {
			continue
		}
		site.CheckType = CheckType(checkType)
		sites = append(sites, site)
	}
	return sites, nil
}

func (s *PostgresStorage) DeleteSite(id string) error {
	_, err := s.db.Exec(`DELETE FROM sites WHERE id=$1`, id)
	if err != nil {
		log.Printf("[ERROR] Postgres DeleteSite: %v", err)
	}
	return err
}

func (s *PostgresStorage) AddCheckHistory(h SiteCheckHistory) error {
	_, err := s.db.Exec(`INSERT INTO check_history (site_id, checked_at, duration_ms, http_code, success, error) VALUES ($1, $2, $3, $4, $5, $6)`,
		h.SiteID, h.CheckedAt, h.Duration, h.HTTPCode, h.Success, h.Error)
	if err != nil {
		log.Printf("[ERROR] Postgres AddCheckHistory: %v", err)
	}
	return err
}

func (s *PostgresStorage) ListCheckHistory(siteID string, limit int) ([]SiteCheckHistory, error) {
	rows, err := s.db.Query(`SELECT site_id, checked_at, duration_ms, http_code, success, error FROM check_history WHERE site_id=$1 ORDER BY checked_at DESC LIMIT $2`, siteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := []SiteCheckHistory{}
	for rows.Next() {
		h := SiteCheckHistory{}
		if err := rows.Scan(&h.SiteID, &h.CheckedAt, &h.Duration, &h.HTTPCode, &h.Success, &h.Error); err != nil {
			continue
		}
		history = append(history, h)
	}
	return history, nil
}

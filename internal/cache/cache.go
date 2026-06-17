package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Service interface {
	Get(trackKey string) (string, bool)
	Set(trackKey, roast string) error
	ClearAll() error
	DeleteEntry(trackKey string) error
	ListAllEntries() ([]string, error) // Updated signature to clean up printing loops
	Close() error
}

type TrackCache struct {
	db *sql.DB
}

type NopCache struct{}

func (n *NopCache) Get(k string) (string, bool) { return "", false }
func (n *NopCache) Set(k, r string) error       { return nil }
func (n *NopCache) ClearAll() error             { return nil }
func (n *NopCache) DeleteEntry(k string) error  { return nil }
func (n *NopCache) Close() error               { return nil }
func (n *NopCache) ListAllEntries() ([]string, error) { return nil, nil }

func GetDatabasePath() (string, error) {
	baseDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(baseDir, "trst")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "roasts.db"), nil
}

func NewTrackCache(dbPath string) (*TrackCache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS roast_cache (
		track_key TEXT PRIMARY KEY,
		roast_output TEXT,
		created_at INTEGER
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return &TrackCache{db: db}, nil
}

func (c *TrackCache) Close() error {
	return c.db.Close()
}

func (c *TrackCache) Get(trackKey string) (string, bool) {
	var roast string
	var createdAt int64

	query := `SELECT roast_output, created_at FROM roast_cache WHERE track_key = ?`
	err := c.db.QueryRow(query, trackKey).Scan(&roast, &createdAt)
	if err != nil {
		return "", false
	}

	if time.Now().Unix()-createdAt > 7*24*60*60 {
		_ = c.DeleteEntry(trackKey)
		return "", false
	}

	return roast, true
}

func (c *TrackCache) Set(trackKey, roast string) error {
	query := `INSERT OR REPLACE INTO roast_cache (track_key, roast_output, created_at) VALUES (?, ?, ?);`
	_, err := c.db.Exec(query, trackKey, roast, time.Now().Unix())
	return err
}

func (c *TrackCache) DeleteEntry(trackKey string) error {
	_, err := c.db.Exec(`DELETE FROM roast_cache WHERE track_key = ?`, trackKey)
	return err
}

func (c *TrackCache) ClearAll() error {
	_, err := c.db.Exec(`DELETE FROM roast_cache`)
	return err
}

func (c *TrackCache) ListAllEntries() ([]string, error) {
	query := `SELECT track_key FROM roast_cache ORDER BY created_at DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err == nil {
			results = append(results, key)
		}
	}
	return results, nil
}

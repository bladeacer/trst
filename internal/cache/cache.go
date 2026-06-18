package cache

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type TrackMeta struct {
	Genre     string
	BPM       int
	CreatedAt int64
}

// CacheEntry marries the key identifier with its structured payload metadata
type CacheEntry struct {
	TrackKey string
	Meta     TrackMeta
}

type Service interface {
	GetTrackMeta(trackKey string) (*TrackMeta, bool)
	SetTrackMeta(trackKey string, genre string, bpm int) error
	ClearAll() error
	DeleteEntry(trackKey string) error
	ListAllEntries() ([]CacheEntry, error) // Updated signature
	Close() error
}

type TrackCache struct {
	db *sql.DB
}

type NopCache struct{}

func (n *NopCache) GetTrackMeta(k string) (*TrackMeta, bool)                 { return nil, false }
func (n *NopCache) SetTrackMeta(trackKey string, g string, b int) error      { return nil }
func (n *NopCache) ClearAll() error                                           { return nil }
func (n *NopCache) DeleteEntry(k string) error                                { return nil }
func (n *NopCache) Close() error                                              { return nil }
func (n *NopCache) ListAllEntries() ([]CacheEntry, error)                        { return nil, nil }

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

	// Migrated schema to track pure musicology parameters
	schema := `
	CREATE TABLE IF NOT EXISTS track_metadata_cache (
		track_key TEXT PRIMARY KEY,
		genre TEXT,
		bpm INTEGER,
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

func (c *TrackCache) GetTrackMeta(trackKey string) (*TrackMeta, bool) {
	var meta TrackMeta
	query := `SELECT genre, bpm, created_at FROM track_metadata_cache WHERE track_key = ?`
	err := c.db.QueryRow(query, trackKey).Scan(&meta.Genre, &meta.BPM, &meta.CreatedAt)
	if err != nil {
		return nil, false
	}

	// Cache expiration fallback (7 days TTL)
	if time.Now().Unix()-meta.CreatedAt > 7*24*60*60 {
		_ = c.DeleteEntry(trackKey)
		return nil, false
	}

	return &meta, true
}

func (c *TrackCache) SetTrackMeta(trackKey string, genre string, bpm int) error {
	query := `INSERT OR REPLACE INTO track_metadata_cache (track_key, genre, bpm, created_at) VALUES (?, ?, ?, ?);`
	_, err := c.db.Exec(query, trackKey, genre, bpm, time.Now().Unix())
	return err
}

func (c *TrackCache) DeleteEntry(trackKey string) error {
	_, err := c.db.Exec(`DELETE FROM track_metadata_cache WHERE track_key = ?`, trackKey)
	return err
}

func (c *TrackCache) ClearAll() error {
	_, err := c.db.Exec(`DELETE FROM track_metadata_cache`)
	return err
}

// ListAllEntries scans and constructs the full entry values sorted by insertion history
func (c *TrackCache) ListAllEntries() ([]CacheEntry, error) {
	query := `SELECT track_key, genre, bpm, created_at FROM track_metadata_cache ORDER BY created_at DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CacheEntry
	for rows.Next() {
		var entry CacheEntry
		err := rows.Scan(
			&entry.TrackKey,
			&entry.Meta.Genre,
			&entry.Meta.BPM,
			&entry.Meta.CreatedAt,
		)
		if err == nil {
			results = append(results, entry)
		}
	}
	return results, nil
}

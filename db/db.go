package db

import (
	"database/sql"
	"encoding/json"
	"github.com/anacrolix/dms/ffmpeg"
	"strings"
)

type Database struct {
	base *sql.DB
}

func Open(dbName string) (*Database, error) {
	dbFile, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	if err = updateVersion(dbFile); err != nil {
		return nil, err
	}
	return &Database{dbFile}, nil
}

func (db Database) Get(hash string) (*Metadata, error) {
	result := &Metadata{FFmpegInfo: &ffmpeg.Info{}}
	ffmpegRaw := make([]byte, 0)
	err := db.base.QueryRow(`SELECT title, ffmpegInfo FROM metadata WHERE hash = ?`, hash).Scan(
		&result.Title, &ffmpegRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal([]byte(ffmpegRaw), result.FFmpegInfo)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db Database) Set(hash string, metadata *Metadata) error {
	ffmpegInfo, err := json.Marshal(metadata.FFmpegInfo)
	if err != nil {
		panic(err)
	}
	_, err = db.base.Exec(`INSERT INTO metadata (hash, title, ffmpegInfo) VALUES (?, ?, ?)`,
		hash, metadata.Title, ffmpegInfo)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) GetThumbnail(hash, thumbType string) ([]byte, error) {
	thumbValue := make([]byte, 0)
	err := db.base.QueryRow(`SELECT thumbValue FROM thumbnails WHERE hash = ? AND thumbType = ?`, hash,
		strings.ToLower(thumbType)).Scan(&thumbValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return thumbValue, nil
}

func (db Database) SaveThumbnail(hash, thumbType string, thumb []byte) error {
	_, err := db.base.Exec(`INSERT OR REPLACE INTO thumbnails (hash, thumbType, thumbValue) VALUES (?, ?, ?);`,
		hash, strings.ToLower(thumbType), thumb)
	if err != nil {
		return err
	}
	return nil
}

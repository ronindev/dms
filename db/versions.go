package db

import (
	"database/sql"
	"errors"
	"strconv"
)

func updateVersion(db *sql.DB) error {
	dbVersion, err := getDBVersion(db)
	if err != nil {
		return errors.New("Error checking current DB version number: " + err.Error())
	}

	if dbVersion == 0 {
		if err := updateDBVersion(db, dbVersion+1); err != nil {
			return err
		}
		dbVersion++
	}
	if dbVersion == 1 {
		_, err := db.Exec(`
CREATE TABLE metadata (hash TEXT PRIMARY KEY, title TEXT, jpegThumb BLOB, pngThumb BLOB, ffmpegInfo BLOB);
			`)
		if err != nil {
			return err
		}
		if err := updateDBVersion(db, dbVersion+1); err != nil {
			return err
		}
		dbVersion++
	}
	if dbVersion == 2 {
		_, err := db.Exec(`
BEGIN TRANSACTION;
CREATE TABLE thumbnails (
  hash TEXT,
  thumbType TEXT,
  thumbValue BLOB,
  PRIMARY KEY (hash, thumbType)
);
INSERT INTO thumbnails
SELECT
	hash,
	'jpeg' AS thumbType,
	jpegThumb AS thumbValue
FROM
	metadata
WHERE
	jpegThumb IS NOT NULL
UNION
SELECT
	hash,
	'png' AS thumbType,
	pngThumb AS thumbValue
FROM
	metadata
WHERE
	pngThumb IS NOT NULL;

CREATE TABLE metadata_b (hash TEXT PRIMARY KEY, title TEXT, ffmpegInfo BLOB);
INSERT INTO metadata_b SELECT hash, title, ffmpegInfo FROM metadata;
DROP TABLE metadata;
ALTER TABLE metadata_b RENAME TO metadata;
COMMIT;
			`)
		if err != nil {
			return err
		}
		if err := updateDBVersion(db, dbVersion+1); err != nil {
			return err
		}
		dbVersion++
	}

	return nil
}

func updateDBVersion(db *sql.DB, v int) error {
	err := setDBVersion(db, v)
	if err != nil {
		return errors.New("Error updating current DB version number: " + err.Error())
	}
	return nil
}

func getDBVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func setDBVersion(db *sql.DB, v int) error {
	_, err := db.Exec(`PRAGMA user_version = ` + strconv.Itoa(v))
	if err != nil {
		return err
	}
	return nil
}

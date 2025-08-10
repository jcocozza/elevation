package db

import (
	"context"
	"database/sql"
	"elevation/internal/hgt"
	"fmt"

	_ "modernc.org/sqlite"
)

func setupDB(db *sql.DB) error {
	_, err := db.Exec("PRAGMA journal_mode = WAL;")
	return err
}

func createSQLiteDB(path string, readOnly bool) (*sql.DB, error) {
	if readOnly {
		path = fmt.Sprintf("%s?mode=ro", path)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := setupDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

const schema string = `
create table if not exists srtm (
	latitude real,
	longitude real,
	elevation real,
	primary key (latitude, longitude)
);`

type ElevationDB interface {
	CreateRecord(ctx context.Context, lat float64, lng float64, elevation float64) error
	// for bulk loading
	CreateRecords(ctx context.Context, records []hgt.HGTRecord) error
	ReadNearestNeighbor(ctx context.Context, lat float64, lng float64) (hgt.HGTRecord, error)
	ReadFourNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([4]hgt.HGTRecord, error)
	ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([16]hgt.HGTRecord, error)
}

func NewElevationDB(path string, readOnly bool) (ElevationDB, error) {
	db, err := createSQLiteDB(path, readOnly)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("error failed to execute schema: %v", err)
	}
	return &ElevationSQLiteDB{db}, nil
}

// Implements ElevationDB
type ElevationSQLiteDB struct {
	*sql.DB
}

func (db *ElevationSQLiteDB) CreateRecord(ctx context.Context, lat float64, lng float64, elevation float64) error {
	q := "insert into srtm (latitude, longitude, elevation) values (?,?,?);"
	_, err := db.ExecContext(ctx, q, lat, lng, elevation)
	return err
}

func (db *ElevationSQLiteDB) CreateRecords(ctx context.Context, records []hgt.HGTRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "insert into srtm (latitude, longitude, elevation) values (?,?,?);")
	if err != nil {
		return err
	}
	for _, record := range records {
		_, err := stmt.ExecContext(ctx, record.Latitude, record.Longitude, record.Elevation)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *ElevationSQLiteDB) ReadNearestNeighbor(ctx context.Context, lat float64, lng float64) (hgt.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
order by (latitude- ?) * (latitude - ?) + (longitude - ?) * (longitude - ?)
limit 1;`
	row := db.QueryRowContext(ctx, q, lat, lat, lng, lng)

	var latitude float64
	var longitude float64
	var elevation float64
	err := row.Scan(
		&latitude,
		&longitude,
		&elevation,
	)
	if err != nil {
		return hgt.HGTRecord{}, err
	}
	return hgt.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elevation}, nil
}
func (db *ElevationSQLiteDB) ReadFourNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([4]hgt.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ? - ?
and latitude <= ? + ?
and longitude >= ? - ?
and longitude <= ? + ?
order by latitude, longitude;`

	records := [4]hgt.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, spacing, lat, spacing, lng, spacing, lng, spacing)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elevation float64
		err := rows.Scan(
			&latitude,
			&longitude,
			&elevation,
		)
		if err != nil {
			return records, err
		}
		records[i] = hgt.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elevation}
		i++
	}
	return records, nil

}
func (db *ElevationSQLiteDB) ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([16]hgt.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ?  - 2 * ?
and latitude <= ? + 2 * ?
and longitude >= ? - 2 * ?
and longitude <= ? + 2 * ?
order by latitude, longitude;`

	records := [16]hgt.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, spacing, lat, spacing, lng, spacing, lng, spacing)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elevation float64
		err := rows.Scan(
			&latitude,
			&longitude,
			&elevation,
		)
		if err != nil {
			return records, err
		}
		if i > 15 {
			return records, fmt.Errorf("%d results returned, expected at most 16", i+1)
		}
		records[i] = hgt.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elevation}
		i++
	}
	return records, nil
}

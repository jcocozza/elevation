package db

import (
	"context"
	"database/sql"
	"elevation/internal/hgt"
	"fmt"
)

func createSQLiteDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

const schema string = `
create table srtm (
	latitude real,
	longitude real,
	elevation integer,
	primary key (latitude, longitude)
);`

type ElevationDB interface {
	CreateRecord(ctx context.Context, lat float64, lng float64, elevation int) error
	ReadNearestNeighbor(ctx context.Context, lat float64, lng float64) (hgt.HGTRecord, error)
	ReadFourNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([4]hgt.HGTRecord, error)
	ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing) ([16]hgt.HGTRecord, error)
}

// Implements ElevationDB
type ElevationSQLiteDB struct {
	*sql.DB
}

func (db *ElevationSQLiteDB) CreateRecord(ctx context.Context, lat float64, lng float64, elevation int) error {
	q := "insert into srtm (latitude, longitude, elevation) values (?,?,?);"
	_, err := db.ExecContext(ctx, q, lat, lng, elevation)
	return err
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
	var elevation int
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
func (db *ElevationSQLiteDB) ReadFourNeighbors(ctx context.Context, lat float64, lng float64) ([4]hgt.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ?  - :spacing
and latitude <= ? + :spacing
and longitude >= ? - :spacing
and longitude <= ? + :spacing
order by latitude, longitude;`

	records := [4]hgt.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, lat, lng, lng)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elevation int
		err := rows.Scan(
			&latitude,
			&longitude,
			&elevation,
		)
		if err != nil {
			return records, err
		}
		if i > 3 {
			return records, fmt.Errorf("%d results returned, expected at most 4", i+1)
		}
		records[i] = hgt.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elevation}
		i++
	}
	return records, nil

}
func (db *ElevationSQLiteDB) ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64) ([16]hgt.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ?  - 2 * :spacing
and latitude <= ? + 2 * :spacing
and longitude >= ? - 2 * :spacing
and longitude <= ? + 2 * :spacing
order by latitude, longitude;`

	records := [16]hgt.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, lat, lng, lng)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elevation int
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

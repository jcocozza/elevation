package db

import (
	"context"
	"database/sql"
	"elevation"
	"fmt"

	_ "modernc.org/sqlite"
)

// repository for interacting with the generated sqlite db
type ElevationDB interface {
	// add a single record
	CreateRecord(ctx context.Context, lat float64, lng float64, elevation float64) error
	// for bulk loading
	//
	// runs a single transaction for all records
	CreateRecords(ctx context.Context, records []elevation.HGTRecord) error
	// copy tmp table over to final table
	// add indexes, and delete tmp table
	CreateFinalTable(ctx context.Context) error
	// return the closest record to the passed lat,lng
	ReadNearestNeighbor(ctx context.Context, lat float64, lng float64) (elevation.HGTRecord, error)
	// return the four closest records to the passed lat,lng
	ReadFourNeighbors(ctx context.Context, lat float64, lng float64, spacing elevation.Spacing) ([4]elevation.HGTRecord, error)
	// return the sixteen closest records to the passed lat,lng
	ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64, spacing elevation.Spacing) ([16]elevation.HGTRecord, error)
}

func setupDB(db *sql.DB) error {
	_, err := db.Exec("PRAGMA journal_mode = OFF;")
	if err != nil {
		return err
	}
	_, err = db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		return err
	}
	_, err = db.Exec("PRAGMA cache_size = 1000000;")
	if err != nil {
		return err
	}
	//_, err = db.Exec("PRAGMA locking_mode = EXCLUSIVE;")
	//if err != nil {return err}
	_, err = db.Exec("PRAGMA temp_store = MEMORY;")
	if err != nil {
		return err
	}
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

// insert optimized
//
// this will be dropped after it is created
const tmpSchema string = `
create table if not exists tmp_srtm (
	latitude real,
	longitude real,
	elevation real
);`

// read/space optimized
const schema string = `
create table if not exists srtm (
	latitude real,
	longitude real,
	elevation real,
	primary key (latitude, longitude)
) without rowid;`

// use readOnly when serving the data as nothing should be written to the db
//
// will create a sqlite db with PRAGMA journal_mode = WAL
func NewElevationDB(path string, readOnly bool) (ElevationDB, error) {
	db, err := createSQLiteDB(path, readOnly)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(tmpSchema)
	if err != nil {
		return nil, fmt.Errorf("error failed to execute schema: %v", err)
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
	q := "insert into tmp_srtm (latitude, longitude, elevation) values (?,?,?);"
	_, err := db.ExecContext(ctx, q, lat, lng, elevation)
	return err
}

func (db *ElevationSQLiteDB) CreateRecords(ctx context.Context, records []elevation.HGTRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "insert into tmp_srtm (latitude, longitude, elevation) values (?,?,?);")
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

func (db *ElevationSQLiteDB) CreateFinalTable(ctx context.Context) error {
	//tx, err := db.BeginTx(ctx, nil)
	//if err != nil {
	//	return err
	//}
	//defer tx.Rollback()

	q1 := `insert into srtm select * from tmp_srtm order by latitude, longitude;`
	_, err := db.ExecContext(ctx, q1)
	if err != nil {
		return err
	}
	q2 := `drop table tmp_srtm;`
	_, err = db.ExecContext(ctx, q2)
	if err != nil {
		return err
	}
	q3 := `VACUUM;` // this can dramatically reduce overall db size
	_, err = db.ExecContext(ctx, q3)
	if err != nil {
		return err
	}
	return nil //tx.Commit()
}

func (db *ElevationSQLiteDB) ReadNearestNeighbor(ctx context.Context, lat float64, lng float64) (elevation.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
order by (latitude- ?) * (latitude - ?) + (longitude - ?) * (longitude - ?)
limit 1;`
	row := db.QueryRowContext(ctx, q, lat, lat, lng, lng)

	var latitude float64
	var longitude float64
	var elev float64
	err := row.Scan(
		&latitude,
		&longitude,
		&elev,
	)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	return elevation.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elev}, nil
}
func (db *ElevationSQLiteDB) ReadFourNeighbors(ctx context.Context, lat float64, lng float64, spacing elevation.Spacing) ([4]elevation.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ? - ?
and latitude <= ? + ?
and longitude >= ? - ?
and longitude <= ? + ?
order by latitude, longitude;`

	records := [4]elevation.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, spacing, lat, spacing, lng, spacing, lng, spacing)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elev float64
		err := rows.Scan(
			&latitude,
			&longitude,
			&elev,
		)
		if err != nil {
			return records, err
		}
		records[i] = elevation.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elev}
		i++
	}
	return records, nil

}
func (db *ElevationSQLiteDB) ReadSixteenNeighbors(ctx context.Context, lat float64, lng float64, spacing elevation.Spacing) ([16]elevation.HGTRecord, error) {
	q := `
select latitude, longitude, elevation
from srtm
where latitude >= ?  - 2 * ?
and latitude <= ? + 2 * ?
and longitude >= ? - 2 * ?
and longitude <= ? + 2 * ?
order by latitude, longitude;`

	records := [16]elevation.HGTRecord{}
	rows, err := db.QueryContext(ctx, q, lat, spacing, lat, spacing, lng, spacing, lng, spacing)
	if err != nil {
		return records, err
	}

	i := 0
	for rows.Next() {
		var latitude float64
		var longitude float64
		var elev float64
		err := rows.Scan(
			&latitude,
			&longitude,
			&elev,
		)
		if err != nil {
			return records, err
		}
		if i > 15 {
			return records, fmt.Errorf("%d results returned, expected at most 16", i+1)
		}
		records[i] = elevation.HGTRecord{Latitude: latitude, Longitude: longitude, Elevation: elev}
		i++
	}
	return records, nil
}

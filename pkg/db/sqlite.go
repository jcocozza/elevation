package db

import (
	"context"
	"database/sql"
	"elevation"
	"fmt"

	_ "modernc.org/sqlite"
)

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
func newElevationSQLiteDB(path string, readOnly bool) (ElevationDB, error) {
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
	return nil, nil
	//return nil, &ElevationSQLiteDB{db}, nil
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

func CreateRecords(db *ElevationSQLiteDB, records []elevation.HGTRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("insert into tmp_srtm (latitude, longitude, elevation) values (?,?,?);")
	if err != nil {
		return err
	}
	for _, record := range records {
		_, err := stmt.Exec(record.Latitude, record.Longitude, record.Elevation)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func CreateFinalTable(db *ElevationSQLiteDB) error {
	//tx, err := db.BeginTx(ctx, nil)
	//if err != nil {
	//	return err
	//}
	//defer tx.Rollback()

	q1 := `insert into srtm select * from tmp_srtm order by latitude, longitude;`
	_, err := db.Exec(q1)
	if err != nil {
		return err
	}
	q2 := `drop table tmp_srtm;`
	_, err = db.Exec(q2)
	if err != nil {
		return err
	}
	q3 := `VACUUM;` // this can dramatically reduce overall db size
	_, err = db.Exec(q3)
	if err != nil {
		return err
	}
	return nil //tx.Commit()
}

func (db *ElevationSQLiteDB) ReadElevation(ctx context.Context, resolution elevation.Resolution, lat float64, lng float64) (elevation.HGTRecord, error) {
	q := `
select avg(elevation)
from (
select latitude, longitude, elevation
from srtm
order by (latitude- ?) * (latitude - ?) + (longitude - ?) * (longitude - ?)
limit 4);`
	row := db.QueryRowContext(ctx, q, lat, lat, lng, lng)

	var elev float64
	err := row.Scan(
		&elev,
	)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	return elevation.HGTRecord{Latitude: lat, Longitude: lng, Elevation: elev}, nil
}

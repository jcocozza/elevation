package db

import (
	"archive/zip"
	"context"
	"elevation"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const ext string = ".SRTMGL1.hgt.zip"

// TODO: eventually we should implement an in memory very that keeps the
// 100 most recent files in memory
type ElevationFsDB struct {
	root string
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return false
}

func newElevationFsDb(root string) (ElevationDB, error) {
	if !pathExists(root) {
		return nil, fmt.Errorf("path does not exist: %s", root)
	}
	return &ElevationFsDB{
		root: root,
	}, nil
}

func (db *ElevationFsDB) ReadElevation(ctx context.Context, resolution elevation.Resolution, lat float64, lng float64) (elevation.HGTRecord, error) {
	name := elevation.LatLngToTileName(lat, lng)
	zr, err := zip.OpenReader(db.root + "/" + name + ext)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	elev, err := elevation.GetElevationFromZip(zr, lat, lng, resolution)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	return elevation.HGTRecord{
		Latitude:  lat,
		Longitude: lng,
		Elevation: elev,
	}, nil
}

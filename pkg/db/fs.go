package db

import (
	"archive/zip"
	"context"
	"elevation"
	"fmt"
)

// TODO: eventually we should implement an in memory version that keeps
// some amount of the data in memory
type ElevationFsDB struct {
	root string
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

package db

import (
	"context"
	"elevation"
	"errors"
	"io/fs"
	"os"
)

const ext string = ".SRTMGL1.hgt.zip"

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

// repository for interacting with srtm data on disk
type ElevationDB interface {
	// return the closest record to the passed lat,lng
	ReadElevation(ctx context.Context, resolution elevation.Resolution, lat float64, lng float64) (elevation.HGTRecord, error)
}

func NewElevationDB(path string, readOnly bool, maxMemSize int64) (ElevationDB, error) {
	//return newElevationFsDb(path)
	return newFsMemCacheDb(path, maxMemSize), nil
}

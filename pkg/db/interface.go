package db

import (
	"context"
	"elevation"
)

// repository for interacting with the generated sqlite db
type ElevationDB interface {
	// return the closest record to the passed lat,lng
	ReadElevation(ctx context.Context, resolution elevation.Resolution, lat float64, lng float64) (elevation.HGTRecord, error)
}

func NewElevationDB(path string, readOnly bool) (ElevationDB, error) {
	return newElevationFsDb(path)
}

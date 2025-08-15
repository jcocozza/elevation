package service

import (
	"context"
	"elevation"
	"elevation/pkg/db"
	"fmt"
)

type ElevationService struct {
	db db.ElevationDB
}

func NewElevationService(db db.ElevationDB) *ElevationService {
	return &ElevationService{db}
}

type InterpolationMethod string

const (
	// closest point
	NearestNeighbor InterpolationMethod = "nearest"
	// 4 points
	Bilinear = "bilinear"
	// 16 points
	Bicubic = "bicubic"
)

//func (s *ElevationService) AddRecord(ctx context.Context, lat float64, lng float64, elevation float64) error {
//	return s.db.CreateRecord(ctx, lat, lng, elevation)
//}

// Use the exported InterpolationMethod type
func (s *ElevationService) GetPointElevation(ctx context.Context, lat float64, lng float64, resolution elevation.Resolution, interpolationMethod InterpolationMethod) (elevation.HGTRecord, error) {
	switch interpolationMethod {
	case NearestNeighbor:
		fallthrough
	case Bicubic:
		fallthrough
	case Bilinear:
		return s.db.ReadElevation(ctx, resolution, lat, lng)
	default:
		return elevation.HGTRecord{}, fmt.Errorf("invalid interpolation method: %s", interpolationMethod)
	}
}

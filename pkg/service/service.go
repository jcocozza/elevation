package service

import (
	"context"
	"elevation"
	"elevation/pkg/db"
	"fmt"
	"math"
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

func (s *ElevationService) GetPointsElevation(ctx context.Context, points [][]float64, resolution elevation.Resolution, interpolationMethod InterpolationMethod) ([]elevation.HGTRecord, error) {
	switch interpolationMethod {
	case NearestNeighbor:
		fallthrough
	case Bicubic:
		fallthrough
	case Bilinear:
		elevations := make([]elevation.HGTRecord, len(points))
		for i, pt := range points {
			if len(pt) != 2 {
				return nil, fmt.Errorf("invalid coord: (%v)", pt)
			}
			rec, err := s.db.ReadElevation(ctx, resolution, pt[0], pt[1])
			if err != nil {
				return nil, err
			}
			elevations[i] = rec
		}
		return elevations, nil
	default:
		return nil, fmt.Errorf("invalid interpolation method: %s", interpolationMethod)
	}
}

// return gain,loss,grade
func (s *ElevationService) ComputeGainLossGrade(ctx context.Context, coords []elevation.HGTRecord) (float64, float64) {
	var gainTotal float64
	var lossTotal float64
	for i := 1; i < len(coords); i++ {
		elevDelta := coords[i].Elevation - coords[i-1].Elevation
		if elevDelta > 0 {
			gainTotal += elevDelta
		} else {
			lossTotal += elevDelta
		}
	}
	return gainTotal, lossTotal
}

const earthRadius = 6371000 // meters

// TODO: this is a rough approximation
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	lat1 = toRad(lat1)
	lat2 = toRad(lat2)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func (s *ElevationService) ComputeNetGrade(ctx context.Context, first elevation.HGTRecord, last elevation.HGTRecord) float64 {
	totalDistance := haversine(first.Latitude, first.Longitude, last.Latitude, last.Longitude)
	delta := last.Elevation - first.Elevation
	return delta / totalDistance
}

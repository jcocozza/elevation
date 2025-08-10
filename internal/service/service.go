package service

import (
	"context"
	"elevation/internal/db"
	"elevation/internal/hgt"
	"fmt"
	"sort"
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

func (s *ElevationService) AddRecord(ctx context.Context, lat float64, lng float64, elevation float64) error {
	return s.db.CreateRecord(ctx, lat, lng, elevation)
}

func (s *ElevationService) GetNearestNeighbor(ctx context.Context, lat float64, lng float64) (hgt.HGTRecord, error) {
	return hgt.HGTRecord{}, nil
}

func (s *ElevationService) GetFourCorners() ([4]hgt.HGTRecord, error) {
	records := [4]hgt.HGTRecord{}
	return records, nil
}

func (s *ElevationService) GetSixteenPoints() ([16]hgt.HGTRecord, error) {
	records := [16]hgt.HGTRecord{}
	return records, nil
}

func bilinearInterpolation(lat float64, lng float64, records [4]hgt.HGTRecord) float64 {
	points := records[:]

	// Find the bounding coordinates
	minLat, maxLat := points[0].Latitude, points[0].Latitude
	minLng, maxLng := points[0].Longitude, points[0].Longitude

	for _, p := range points {
		if p.Latitude < minLat {
			minLat = p.Latitude
		}
		if p.Latitude > maxLat {
			maxLat = p.Latitude
		}
		if p.Longitude < minLng {
			minLng = p.Longitude
		}
		if p.Longitude > maxLng {
			maxLng = p.Longitude
		}
	}

	// Find the four corners
	var q11, q12, q21, q22 hgt.HGTRecord
	for _, p := range points {
		if p.Latitude == minLat && p.Longitude == minLng {
			q11 = p
		} // bottom-left
		if p.Latitude == maxLat && p.Longitude == minLng {
			q12 = p
		} // top-left
		if p.Latitude == minLat && p.Longitude == maxLng {
			q21 = p
		} // bottom-right
		if p.Latitude == maxLat && p.Longitude == maxLng {
			q22 = p
		} // top-right
	}

	x1, x2 := minLng, maxLng
	y1, y2 := minLat, maxLat

	denom := (x2 - x1) * (y2 - y1)
	if denom == 0 {
		return float64(q11.Elevation) // fallback to any point
	}

	eq11 := float64(q11.Elevation)
	eq21 := float64(q21.Elevation)
	eq12 := float64(q12.Elevation)
	eq22 := float64(q22.Elevation)

	return eq11*(x2-lng)*(y2-lat)/denom +
		eq21*(lng-x1)*(y2-lat)/denom +
		eq12*(x2-lng)*(lat-y1)/denom +
		eq22*(lng-x1)*(lat-y1)/denom
}

func catmullRom(p0, p1, p2, p3, t float64) float64 {
	t2 := t * t
	t3 := t2 * t

	return 0.5 * (2*p1 +
		(-p0+p2)*t +
		(2*p0-5*p1+4*p2-p3)*t2 +
		(-p0+3*p1-3*p2+p3)*t3)
}

func bicubicInterpolation(lat float64, lng float64, records [16]hgt.HGTRecord) float64 {
	points := records[:]
	sort.Slice(points, func(i int, j int) bool {
		if points[i].Latitude == points[j].Latitude {
			return points[i].Longitude < points[j].Longitude
		}
		return points[i].Latitude < points[j].Latitude
	})

	// Create grid - assuming sorted points form regular 4x4 grid
	grid := make([][]float64, 4)
	for i := range 4 {
		grid[i] = make([]float64, 4)
		for j := range 4 {
			grid[i][j] = float64(points[i*4+j].Elevation)
		}
	}

	// Get actual grid bounds
	minLat, maxLat := points[0].Latitude, points[15].Latitude
	minLng, maxLng := points[0].Longitude, points[3].Longitude

	// Normalize coordinates
	tx := (lng - minLng) / (maxLng - minLng)
	ty := (lat - minLat) / (maxLat - minLat)

	// Clamp to valid range
	if tx < 0 {
		tx = 0
	}
	if tx > 1 {
		tx = 1
	}
	if ty < 0 {
		ty = 0
	}
	if ty > 1 {
		ty = 1
	}

	// Interpolate along rows first
	col := make([]float64, 4)
	for row := range 4 {
		col[row] = catmullRom(grid[row][0], grid[row][1], grid[row][2], grid[row][3], tx)
	}

	// Then interpolate along column
	return catmullRom(col[0], col[1], col[2], col[3], ty)
}

// Use the exported InterpolationMethod type
func (s *ElevationService) GetPointElevation(ctx context.Context, lat float64, lng float64, spacing hgt.Spacing, interpolationMethod InterpolationMethod) (hgt.HGTRecord, error) {
	switch interpolationMethod {
	case NearestNeighbor:
		return s.db.ReadNearestNeighbor(ctx, lat, lng)
	case Bilinear:
		records, err := s.db.ReadFourNeighbors(ctx, lat, lng, spacing)
		if err != nil {
			return hgt.HGTRecord{}, err
		}
		elevation := bilinearInterpolation(lat, lng, records)
		return hgt.HGTRecord{Latitude: lat, Longitude: lng, Elevation: elevation}, nil
	case Bicubic:
		records, err := s.db.ReadSixteenNeighbors(ctx, lat, lng, spacing)
		if err != nil {
			return hgt.HGTRecord{}, err
		}
		elevation := bicubicInterpolation(lat, lng, records)
		return hgt.HGTRecord{Latitude: lat, Longitude: lng, Elevation: elevation}, nil
	default:
		return hgt.HGTRecord{}, fmt.Errorf("invalid interpolation method: %s", interpolationMethod)
	}
}

package elevation

// represents the type of interpolation to use
type InterpolationMethod string

// represents the type of interpolation to use
const (
	// closest point
	NearestNeighbor InterpolationMethod = "nearest"
	// 4 points
	Bilinear = "bilinear"
	// 16 points
	Bicubic = "bicubic"
)

type Spacing float64

// constants for spacing
//
// SRTM product	  Approx spacing (arc-seconds)	Decimal degrees
// SRTM1 (≈30 m)	1 arc-second	            1 / 3600 ≈ 0.0002777778°
// SRTM3 (≈90 m)	3 arc-seconds	            3 / 3600 ≈ 0.0008333333°
const (
	SRTM1 Spacing = 1.0 / 3600.0
	SRTM3 Spacing = 3.0 / 3600.0
)

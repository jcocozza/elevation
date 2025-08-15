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

// SRTM product
// SRTM1 (≈30 m)
// SRTM3 (≈90 m)
type Resolution int

func (d Resolution) Gridsize() int {
	return int(d) * int(d)
}

var SRTM1 Resolution = 1201
var SRTM3 Resolution = 3601

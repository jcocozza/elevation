package elevation

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
)

// ensure title name is of the correct format
//
// should look like S15W040
var tileNamePattern = regexp.MustCompile(`^[NS]\d{2}[EW]\d{3}$`)

// return lat, lng, error
//
// tile name is of the form S15W040
func ParseTileName(name string) (int, int, error) {
	if !tileNamePattern.MatchString(name) {
		return -1, -1, fmt.Errorf("invalid title name: %s", name)
	}
	latStr := name[1:3]
	lat, err := strconv.Atoi(latStr)
	if err != nil {
		return -1, -1, err
	}
	lngStr := name[4:7]
	lng, err := strconv.Atoi(lngStr)
	if err != nil {
		return -1, -1, err
	}
	if name[0] == 'S' {
		lat = -lat
	}
	if name[3] == 'W' {
		lng = -lng
	}
	return lat, lng, nil
}

func LatLngToTileName(lat float64, lng float64) string {
	var latPrefix string
	var tileLat float64

	if lat >= 0 {
		latPrefix = "N"
		tileLat = math.Floor(lat)
	} else {
		latPrefix = "S"
		tileLat = math.Floor(lat)
	}

	var lngPrefix string
	var tileLng float64

	if lng >= 0 {
		lngPrefix = "E"
		tileLng = math.Floor(lng)
	} else {
		lngPrefix = "W"
		tileLng = math.Floor(lng)
	}

	return fmt.Sprintf("%s%02d%s%03d", latPrefix, int(math.Abs(tileLat)), lngPrefix, int(math.Abs(tileLng)))
}

// process hgt content
//
// tileLat/tileLng correspond to the data read by the io.Reader
func ProcessHGTFile(r io.Reader, tileLat int, tileLng int) ([]HGTRecord, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	size := int64(len(data))
	totalPoints := int(size / 2)
	var gridSize int
	switch totalPoints {
	case 1201 * 1201:
		gridSize = 1201 // 3-arcsecond data
	case 3601 * 3601:
		gridSize = 3601 // 1-arcsecond data
	default:
		return nil, fmt.Errorf("unexpected file size: %d bytes (%d points)", size, totalPoints)
	}
	elevationData := make([]int16, totalPoints)
	for i := range totalPoints {
		elevationData[i] = int16(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
	}
	// Calculate step size based on grid size
	var step float64
	if gridSize == 1201 {
		step = 1.0 / 1200.0 // 3-arcsecond = 1/1200 degree
	} else {
		step = 1.0 / 3600.0 // 1-arcsecond = 1/3600 degree
	}
	records := make([]HGTRecord, gridSize*gridSize)
	i := 0
	for row := range gridSize {
		for col := range gridSize {
			currentLat := float64(tileLat) + 1.0 - (float64(row) * step)
			currentLng := float64(tileLng) + (float64(col) * step)
			// Get elevation value
			index := row*gridSize + col
			elevation := elevationData[index]
			// Skip NODATA values (typically -32768)
			if elevation == -32768 {
				continue
			}
			records[i] = HGTRecord{Latitude: currentLat, Longitude: currentLng, Elevation: float64(elevation)}
			i++
		}
	}
	return records, nil
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func pixelElevationIdx(x int, y int, r Resolution) (int, error) {
	index := y*int(r) + x
	if index < 0 || index >= r.Gridsize() {
		return -1, fmt.Errorf("pixel out of bounds: (%d,%d)", x, y)
	}
	return index, nil
}

// extract elevation from the data of a hgt file
//
// this does bilinear interpolation
func GetElevation(data []int16, resolution Resolution, tileLat int, tileLng int, lat float64, lng float64) (float64, error) {
	tileLatF := float64(tileLat)
	tileLngF := float64(tileLng)
	if lat < tileLatF || lat >= tileLatF+1.0 || lng < tileLngF || lng >= tileLngF+1.0 {
		return 0, fmt.Errorf("coordinates (%.6f, %.6f) are outside HGT file bounds (%d, %d)", lat, lng, tileLat, tileLng)
	}

	pixelSize := 1.0 / float64(resolution-1)
	x := (lng - tileLngF) / pixelSize
	y := (tileLatF + 1.0 - lat) / pixelSize

	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	x0 = clamp(x0, 0, int(resolution)-2)
	y0 = clamp(y0, 0, int(resolution)-2)
	x1 = x0 + 1
	y1 = y0 + 1

	i00, err := pixelElevationIdx(x0, y0, resolution)
	if err != nil {
		return 0, err
	}
	i10, err := pixelElevationIdx(x1, y0, resolution)
	if err != nil {
		return 0, err
	}
	i01, err := pixelElevationIdx(x0, y1, resolution)
	if err != nil {
		return 0, err
	}
	i11, err := pixelElevationIdx(x1, y1, resolution)
	if err != nil {
		return 0, err
	}
	// Get the 4 corner elevations
	z00 := float64(data[i00])
	z10 := float64(data[i10])
	z01 := float64(data[i01])
	z11 := float64(data[i11])

	voidValue := -32768.0
	if z00 == voidValue || z10 == voidValue || z01 == voidValue || z11 == voidValue {
		// Simple approach: return 0 for void data
		// You might want to implement more sophisticated void filling
		return 0, nil
	}
	fx := x - float64(x0)
	fy := y - float64(y0)
	// Bilinear interpolation
	// First interpolate in x direction
	z0 := z00*(1-fx) + z10*fx
	z1 := z01*(1-fx) + z11*fx
	// Then interpolate in y direction
	elevation := z0*(1-fy) + z1*fy
	return elevation, nil
}

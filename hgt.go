package elevation

import (
	"encoding/binary"
	"fmt"
	"io"
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

// process hgt content
//
// lat/lng corresponds to the data read by the io.Reader
func ProcessHGT(r io.Reader, lat int, lng int) ([]HGTRecord, error) {
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
			currentLat := float64(lat) + 1.0 - (float64(row) * step)
			currentLng := float64(lng) + (float64(col) * step)
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

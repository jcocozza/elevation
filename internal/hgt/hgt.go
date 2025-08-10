package hgt

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// ensure title name is of the correct format
var tileNamePattern = regexp.MustCompile(`^[NS]\d{2}[EW]\d{3}$`)


type Spacing float64
// constants for spacing
//
// SRTM product	  Approx spacing (arc-seconds)	Decimal degrees
// SRTM1 (≈30 m)	1 arc-second	            1 / 3600 ≈ 0.0002777778°
// SRTM3 (≈90 m)	3 arc-seconds	            3 / 3600 ≈ 0.0008333333°
const (
	SRTM1 Spacing = 1.0/3600.0
	SRTM3 Spacing = 3.0/3600.0
)

type HGTRecord struct {
	Latitude  float64
	Longitude float64
	Elevation int
}

func (r HGTRecord) String() string {
	return fmt.Sprintf("%f %f %d", r.Latitude, r.Longitude, r.Elevation)
}

// return csv string:
//
// lat,lng,elevation
func (r HGTRecord) CSV() []string {
	return []string{
		fmt.Sprintf("%f", r.Latitude),
		fmt.Sprintf("%f", r.Longitude),
		fmt.Sprintf("%d", r.Elevation),
	}
}

type HGTRecords []HGTRecord

func (r HGTRecords) CSV(w io.Writer, header bool) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if header {
		err := writer.Write([]string{"latitude", "longitude", "elevation"})
		if err != nil {
			return err
		}

	}
	for _, record := range r {
		err := writer.Write(record.CSV())
		if err != nil {
			return err
		}
	}
	return writer.Error()
}

// return lat, lng, error
func ParseTileName(name string) (int, int, error) {
	if !tileNamePattern.MatchString(name) {
		return -1, -1, fmt.Errorf("invalid title name")
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
	for i := 0; i < totalPoints; i++ {
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

			records[i] = HGTRecord{Latitude: currentLat, Longitude: currentLng, Elevation: int(elevation)}
			i++
		}
	}
	return records, nil
}

package elevation

import (
	"encoding/csv"
	"fmt"
	"io"
)

type HGTRecord struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`
}

func (r HGTRecord) String() string {
	return fmt.Sprintf("%f %f %f", r.Latitude, r.Longitude, r.Elevation)
}

// return csv string:
//
// lat,lng,elevation
func (r HGTRecord) csv() []string {
	return []string{
		fmt.Sprintf("%f", r.Latitude),
		fmt.Sprintf("%f", r.Longitude),
		fmt.Sprintf("%f", r.Elevation),
	}
}

// write records to w in csv form
func HGTToCSV(w io.Writer, header bool, records []HGTRecord) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if header {
		err := writer.Write([]string{"latitude", "longitude", "elevation"})
		if err != nil {
			return err
		}
	}
	for _, record := range records {
		err := writer.Write(record.csv())
		if err != nil {
			return err
		}
	}
	return writer.Error()
}

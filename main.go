package main

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <hgt_file>")
		fmt.Println("Example: go run main.go N00E006.hgt")
		os.Exit(1)
	}

	hgtFile := os.Args[1]
	csvFile := strings.TrimSuffix(hgtFile, ".hgt") + ".csv"

	err := convertHGTToCSV(hgtFile, csvFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", hgtFile, csvFile)
}

func convertHGTToCSV(hgtFile, csvFile string) error {
	// Parse filename to get coordinates (e.g., N00E006.hgt)
	baseName := strings.TrimSuffix(hgtFile, ".hgt")
	if len(baseName) < 7 {
		return fmt.Errorf("invalid HGT filename format")
	}

	// Extract latitude and longitude from filename
	latStr := baseName[1:3]
	lngStr := baseName[4:7]

	lat, err := strconv.Atoi(latStr)
	if err != nil {
		return fmt.Errorf("error parsing latitude: %v", err)
	}

	lng, err := strconv.Atoi(lngStr)
	if err != nil {
		return fmt.Errorf("error parsing longitude: %v", err)
	}

	// Handle N/S and E/W directions
	if baseName[0] == 'S' {
		lat = -lat
	}
	if baseName[3] == 'W' {
		lng = -lng
	}

	// Open HGT file
	file, err := os.Open(hgtFile)
	if err != nil {
		return fmt.Errorf("error opening HGT file: %v", err)
	}
	defer file.Close()

	// Get file size to determine grid size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}

	fileSize := fileInfo.Size()
	// Each elevation value is 2 bytes (int16)
	totalPoints := int(fileSize / 2)

	// SRTM files are typically 1201x1201 (3601x3601 for 1-arcsec data)
	var gridSize int
	switch totalPoints {
	case 1201 * 1201:
		gridSize = 1201 // 3-arcsecond data
	case 3601 * 3601:
		gridSize = 3601 // 1-arcsecond data
	default:
		return fmt.Errorf("unexpected file size: %d bytes (%d points)", fileSize, totalPoints)
	}

	fmt.Printf("Processing %dx%d grid\n", gridSize, gridSize)

	// Create CSV file
	csvFileHandle, err := os.Create(csvFile)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer csvFileHandle.Close()

	writer := csv.NewWriter(csvFileHandle)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"latitude", "longitude", "elevation"})

	// Calculate step size based on grid size
	var step float64
	if gridSize == 1201 {
		step = 1.0 / 1200.0 // 3-arcsecond = 1/1200 degree
	} else {
		step = 1.0 / 3600.0 // 1-arcsecond = 1/3600 degree
	}

	// Read and convert elevation data
	elevationData := make([]int16, totalPoints)
	err = binary.Read(file, binary.BigEndian, elevationData)
	if err != nil {
		return fmt.Errorf("error reading elevation data: %v", err)
	}

	// Convert to CSV
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			// Calculate lat/lng for this point
			// SRTM data starts from top-left (north-west corner)
			currentLat := float64(lat) + 1.0 - (float64(row) * step)
			currentLng := float64(lng) + (float64(col) * step)

			// Get elevation value
			index := row*gridSize + col
			elevation := elevationData[index]

			// Skip NODATA values (typically -32768)
			if elevation == -32768 {
				continue
			}

			// Write to CSV
			record := []string{
				fmt.Sprintf("%.6f", currentLat),
				fmt.Sprintf("%.6f", currentLng),
				fmt.Sprintf("%d", elevation),
			}
			writer.Write(record)
		}

		// Progress indicator
		if row%100 == 0 {
			fmt.Printf("Progress: %.1f%%\r", float64(row)/float64(gridSize)*100)
		}
	}

	fmt.Printf("Progress: 100.0%%\n")
	return nil
}

package elevation

import (
	"archive/zip"
	"encoding/binary"
	"path/filepath"
	"strings"
)

func parseFilename(name string) string {
	return strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
}

func ProcessZippedHGT(zr *zip.ReadCloser) ([]HGTRecord, error) {
	zf := zr.File[0]
	fname := zr.File[0].Name
	lat, lng, err := ParseTileName(parseFilename(fname))
	if err != nil {
		return nil, err
	}
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ProcessHGTFile(f, lat, lng)
}

// get the elevation of a (lat,lng) pair from a zip
func GetElevationFromZip(zr *zip.ReadCloser, lat float64, lng float64, resolution Resolution) (float64, error) {
	zf := zr.File[0]
	fname := zr.File[0].Name
	tileLat, tileLng, err := ParseTileName(parseFilename(fname))
	if err != nil {
		return 0, err
	}

	f, err := zf.Open()
	if err != nil {
		return 0, err
	}
	defer f.Close()

	data := make([]int16, resolution.Gridsize())
	err = binary.Read(f, binary.BigEndian, data)
	if err != nil {
		return -1, err
	}
	return GetElevation(data, resolution, tileLat, tileLng, lat, lng)
}

package util

import "fmt"

func ValidateLatLng(lat float64, lng float64) error {
	var latMsg string
	if lat < -90 {
		latMsg += fmt.Sprintf("lat is out of bounds: %f < -90", lat)
	} else if lat > 90 {
		latMsg += fmt.Sprintf("lat is out of bounds: %f > 90", lat)
	}

	var lngMsg string
	if lng < -180 {
		lngMsg += fmt.Sprintf("lng is out of bounds: %f < -180", lng)
	} else if lng > 180 {
		lngMsg += fmt.Sprintf("lng is out of bounds: %f > 180", lng)
	}

	if latMsg != "" && lngMsg != "" {
		return fmt.Errorf("%s; %s", latMsg, lngMsg)
	} else if latMsg != "" {
		return fmt.Errorf("%s", latMsg)
	} else if lngMsg != "" {
		return fmt.Errorf("%s", lngMsg)
	}
	return nil
}

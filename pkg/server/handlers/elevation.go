package handlers

import (
	"context"
	"elevation"
	"elevation/pkg/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type ElevationHandler struct {
	s *service.ElevationService
}

func NewElevationHandler(s *service.ElevationService) *ElevationHandler {
	return &ElevationHandler{s: s}
}

// checks for the following query params:
// - latitude
// - longitude
func (h *ElevationHandler) ElevationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//params := r.URL.Query()
		//interpolationMethod := service.InterpolationMethod(params.Get("interpolation"))
		//if interpolationMethod == "" {
		//	interpolationMethod = service.Bilinear
		//}
		latStr := r.PathValue("latitude")
		lngStr := r.PathValue("longitude")

		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to parse latitude: %s", err.Error()), http.StatusBadRequest)
			return
		}
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to parse longitude: %s", err.Error()), http.StatusBadRequest)
			return
		}
		record, err := h.s.GetPointElevation(context.Background(), lat, lng, elevation.SRTM3, elevation.Bilinear)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get elevation: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(record)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode json: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	default:
	}
}

package handlers

import (
	"context"
	"elevation/internal/hgt"
	"elevation/internal/service"
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
// - interpolation (can be nearest, bilinear, bicubic) (default bilinear)
func (h *ElevationHandler) ElevationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		params := r.URL.Query()

		interpolationMethod := service.InterpolationMethod(params.Get("interpolation"))
		if interpolationMethod == "" {
			interpolationMethod = service.Bilinear
		}
		latStr := params.Get("latitude")
		lngStr := params.Get("longitude")

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

		record, err := h.s.GetPointElevation(context.Background(), lat, lng, hgt.SRTM1, interpolationMethod)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get elevation: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(record)
		if err != nil {
			http.Error(w, fmt.Sprintf("faile to encode json: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	default:
	}
}

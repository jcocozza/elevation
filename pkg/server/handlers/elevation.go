package handlers

import (
	"context"
	"elevation"
	"elevation/pkg/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/twpayne/go-polyline"
)

type ElevationHandler struct {
	s *service.ElevationService
}

func NewElevationHandler(s *service.ElevationService) *ElevationHandler {
	return &ElevationHandler{s: s}
}

// checks for the following query params:
// - latitude (lat)
// - longitude (lng)
func (h *ElevationHandler) PointHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//params := r.URL.Query()
		//interpolationMethod := service.InterpolationMethod(params.Get("interpolation"))
		//if interpolationMethod == "" {
		//	interpolationMethod = service.Bilinear
		//}
		params := r.URL.Query()
		latStr := params.Get("lat")
		lngStr := params.Get("lng")

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

type pointsRequest struct {
	Polyline string       `json:"polyline"`
	// this really should be [][2]float64
	Points   [][]float64 `json:"points"`
}

func (h *ElevationHandler) PolylineHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		params := r.URL.Query()
		route := params.Get("route")
		if route == "" {
			http.Error(w, "empty route", http.StatusBadRequest)
			return
		}
		coords, _, err := polyline.DecodeCoords([]byte(route))
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to decode polyline: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(coords)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode json: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read json: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		var req pointsRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal: %s", err.Error()), http.StatusBadRequest)
			return
		}

		var coords [][]float64
		if req.Polyline != "" {
			coords, _, err = polyline.DecodeCoords([]byte(req.Polyline))
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to decode polyline: %s", err.Error()), http.StatusBadRequest)
				return
			}
		} else {
			coords = req.Points
		}

		records, err := h.s.GetPointsElevation(context.TODO(), coords, elevation.SRTM3, elevation.Bilinear)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get data: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(records)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode json: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

package server

import (
	"elevation/pkg/server/handlers"
	"elevation/pkg/service"
	"fmt"
	"net/http"
)

func router(handler *handlers.ElevationHandler) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("alive")) })
	api.HandleFunc("/elevation/points", handler.PolylineHandler)
	api.HandleFunc("/elevation/point", handler.PointHandler)
	return api
}

func Serve(address string, port int, s *service.ElevationService) error {
	h := handlers.NewElevationHandler(s)
	r := router(h)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), r)
	return err
}

package http

import (
	"elevation/internal/http/handlers"
	"elevation/internal/service"
	"fmt"
	"net/http"
)

func router(handler *handlers.ElevationHandler) http.Handler {
	// /api routes
	api := http.NewServeMux()

	api.HandleFunc("/elevation", handler.ElevationHandler)
	return api
}

func Serve(address string, port int, s *service.ElevationService) error {
	h := handlers.NewElevationHandler(s)
	r := router(h)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), r)
	return err
}

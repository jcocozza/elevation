package server

import (
	"elevation/pkg/server/handlers"
	"elevation/pkg/service"
	"fmt"
	"net/http"
)

func router(handler *handlers.ElevationHandler) http.Handler {
	// /api routes
	api := http.NewServeMux()
	api.HandleFunc("/elevation/{latitude}/{longitude}", handler.ElevationHandler)
	return api
}

func Serve(address string, port int, s *service.ElevationService) error {
	h := handlers.NewElevationHandler(s)
	r := router(h)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), r)
	return err
}

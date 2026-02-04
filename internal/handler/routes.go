package handler

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {

	mux.HandleFunc("POST /query", h.HandleUserQuery)
	
}

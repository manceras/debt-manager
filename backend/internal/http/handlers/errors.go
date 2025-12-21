package handlers

import "net/http"

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "endpoint not found")
}

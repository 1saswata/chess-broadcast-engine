package api

import (
	"net/http"
)

func (s *APIServer) SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", s.HandleRegister)
	mux.HandleFunc("POST /login", s.HandleLogin)
	mux.HandleFunc("POST /archive", s.HandleArchive)
	return mux
}

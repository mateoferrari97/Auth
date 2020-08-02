package processor

import (
	"fmt"
	"net/http"
)

func (s *Server) wrapHandler(method, pattern string, handler http.HandlerFunc) {
	s.router.HandleFunc(pattern, handler).Methods(method)
}

func (s *Server) Ping() {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Pong")
	}

	s.wrapHandler(http.MethodGet, "/ping", wrapH)
}

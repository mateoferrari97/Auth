package server

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const defaultPort = "8081"

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	return &Server{router: mux.NewRouter()}
}

func (s *Server) Run(port string) {
	port = configPort(port)

	log.Printf("Listening on port %s", port)

	http.ListenAndServe(":" + port, s.router)
}

func (s *Server) Wrap(method string, pattern string, handler http.HandlerFunc) {
	s.router.HandleFunc(pattern, handler).Methods(method)
}

func configPort(port string) string {
	if port == "" {
		port = defaultPort
		log.Printf("Defaulting to port %s", port)
	}

	if string(port[0]) == ":" {
		port = port[1:]
	}

	return port
}

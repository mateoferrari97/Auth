package processor

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const _defaultPort = "8081"

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

func configPort(port string) string {
	if port == "" {
		port = _defaultPort
		log.Printf("defaulting to port %s", port)
	}

	if string(port[0]) == ":" {
		port = port[1:]
	}

	return port
}

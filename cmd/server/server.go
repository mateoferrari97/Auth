package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mateoferrari97/auth/internal"
)

const defaultPort = "8081"

type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	return &Server{Router: mux.NewRouter()}
}

func (s *Server) Run(port string) error {
	port = configPort(port)

	log.Printf("Listening on port %s", port)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), s.Router)
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (s *Server) Wrap(method string, pattern string, handler HandlerFunc) {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		hErr := handleError(err)
		_ = internal.RespondJSON(w, hErr, hErr.StatusCode)
	}

	s.Router.HandleFunc(pattern, wrapH).Methods(method)
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

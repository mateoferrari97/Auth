package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mateoferrari97/auth/internal"
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

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), s.router); err != nil {
		panic(err)
	}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (s *Server) Wrap(method string, pattern string, handler HandlerFunc) {
	wrapH :=  func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		hErr := HandleError(err)
		_ = internal.RespondJSON(w, hErr, hErr.StatusCode)
	}

	s.router.HandleFunc(pattern, wrapH).Methods(method)
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

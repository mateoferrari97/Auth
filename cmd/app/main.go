package main

import (
	"github.com/mateoferrari97/auth/cmd/app/internal"
	"github.com/mateoferrari97/auth/cmd/server"
)

func main() {
	server := server.NewServer()
	handler := internal.NewHandler(server)
	service := internal.NewService()

	handler.Ping()
	handler.RouteHome(service.Authorize)
	handler.RouteRegister(service.Register)

	server.Run(":8081")
}

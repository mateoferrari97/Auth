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
	handler.RouteMe(service.Authorize)
	handler.RouteRegister(service.Register)
	handler.RouteLoginWithGoogle(service.LoginWithGoogle)
	handler.RouteLoginWithGoogleCallback(service.LoginWithGoogleCallback)
	handler.RouteLogout()

	server.Run(":8081")
}

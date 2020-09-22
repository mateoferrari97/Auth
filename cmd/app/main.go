package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mateoferrari97/auth/cmd/app/internal"
	"github.com/mateoferrari97/auth/cmd/app/internal/client"
	"github.com/mateoferrari97/auth/cmd/server"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	server := server.NewServer()

	repository, err := newUserRepository()
	if err != nil {
		return err
	}

	client := client.NewClient(http.DefaultClient)
	service := internal.NewService(repository, client)
	handler := internal.NewHandler(server)

	handler.Ping()
	handler.RouteMe(service.Authorize)
	handler.RouteRegister(service.Register)
	handler.RouteLoginWithGoogle(service.LoginWithGoogle)
	handler.RouteLoginWithGoogleCallback(service.LoginWithGoogleCallback)
	handler.RouteLogout()

	return server.Run(":8081")
}

func newUserRepository() (internal.Repository, error) {
	dbSettings := fmt.Sprintf("%s:%s@tcp(db:3306)/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
	)

	db, err := sqlx.Connect("mysql", dbSettings)
	if err != nil {
		return nil, fmt.Errorf("instantiating db: %v", err)
	}

	return internal.NewUserRepository(db), nil
}

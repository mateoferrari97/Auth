package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mateoferrari97/auth/cmd/app/internal"
	"github.com/mateoferrari97/auth/cmd/server"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	server := server.NewServer()
	handler := internal.NewHandler(server)
	service := internal.NewService()

	dbSettings := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
	)

	db, err := sqlx.Connect("mysql", dbSettings)
	if err != nil {
		return fmt.Errorf("instantiating db: %v", err)
	}

	log.Printf("DB SUCCESS: %s", db.DriverName())

	handler.Ping()
	handler.RouteMe(service.Authorize)
	handler.RouteRegister(service.Register)
	handler.RouteLoginWithGoogle(service.LoginWithGoogle)
	handler.RouteLoginWithGoogleCallback(service.LoginWithGoogleCallback)
	handler.RouteLogout()

	return server.Run(":8081")
}

package main

import (
	"github.com/mateoferrari97/auth/cmd/processor"
)

func main() {
	s := processor.NewServer()
	s.Ping()

	s.Run(":8081")
}

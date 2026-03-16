package main

import (
	"fmt"
	"os"

	chronolog "github.com/Vp-2306/Chronolog"
	"github.com/Vp-2306/Chronolog/internal/server"
)

func main() {

	engine, err := chronolog.NewEngine("chronolog.wal")
	if err != nil {
		fmt.Println("Error starting engine:", err)
		os.Exit(1)
	}
	defer engine.Close()

	s := server.NewServer(engine, "8080")
	if err := s.Start(); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
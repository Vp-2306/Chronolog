package main

import (
	"fmt"
	"os"

	chronolog "github.com/Vp-2306/Chronolog"
)

func main() {

	engine, err := chronolog.NewEngine("chronolog.wal")
	if err != nil {
		fmt.Println("Error starting engine:", err)
		os.Exit(1)
	}
	defer engine.Close()

	engine.Put([]byte("apple"), []byte("10"))
	engine.Put([]byte("banana"), []byte("20"))
	engine.Put([]byte("cherry"), []byte("30"))

	val, err := engine.Get([]byte("apple"))
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("apple:", string(val))
	}

	val, err = engine.Get([]byte("banana"))
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("banana:", string(val))
	}
}
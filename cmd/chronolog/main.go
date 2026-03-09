package main

import (
	"fmt"
	"os"

	"github.com/Vp-2306/Chronolog/internal/memtable"
	"github.com/Vp-2306/Chronolog/internal/wal"
)

func main() {

	mem := memtable.NewMemTable()

	// check if WAL exists
	if _, err := os.Stat("chronolog.wal"); err == nil {

		fmt.Println("Recovering from WAL...")

		w, _ := wal.NewWAL("chronolog.wal")

		w.Recover(mem)

		w.Close()
	}

	// open WAL for new writes
	w, _ := wal.NewWAL("chronolog.wal")

	w.Append([]byte("PUT apple 10\n"))
	mem.Put([]byte("apple"), []byte("10"))

	val := mem.Get([]byte("apple"))

	fmt.Println("Value:", string(val))
}
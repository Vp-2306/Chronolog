package main

import (
	"fmt"
	"github.com/Vp-2306/Chronolog/internal/memtable"
)

func main() {

	sl := memtable.NewSkipList()

	sl.Insert([]byte("apple"), []byte("10"))
	sl.Insert([]byte("banana"), []byte("20"))

	value := sl.Search([]byte("apple"))

	fmt.Println(string(value))
}
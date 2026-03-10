package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/Vp-2306/Chronolog/internal/memtable"
)

func WriteSSTable(mem *memtable.MemTable) (string, error) {

	filename := fmt.Sprintf("%d.sst", time.Now().UnixNano())

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	iter := mem.NewIterator()

	for iter.HasNext() {

		key, value := iter.Next()

		// skip tombstones
		if value == nil {
			continue
		}

		// write key length
		keyLen := uint32(len(key))
		if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
			return "", err
		}

		// write key
		if _, err := file.Write(key); err != nil {
			return "", err
		}

		// write value length
		valLen := uint32(len(value))
		if err := binary.Write(file, binary.LittleEndian, valLen); err != nil {
			return "", err
		}

		// write value
		if _, err := file.Write(value); err != nil {
			return "", err
		}
	}

	return filename, nil
}
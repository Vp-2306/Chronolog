package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"github.com/Vp-2306/Chronolog/internal/bloom"
	"github.com/Vp-2306/Chronolog/internal/memtable"
)

type SSTableWriter struct{
	filter *bloom.BloomFilter
}


func WriteSSTable(mem *memtable.MemTable) (string, *bloom.BloomFilter, error) {

	filename := fmt.Sprintf("%d.sst", time.Now().UnixNano())

	file, err := os.Create(filename)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	filter := bloom.NewBloomFilter()

	iter := mem.NewIterator()

	for iter.HasNext() {

		key, value := iter.Next()

		// skip tombstones
		if value == nil {
			continue
		}

		//add key to bloom filter
		filter.Add(key)

		// write key length
		keyLen := uint32(len(key))
		if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
			return "", nil, err
		}

		// write key
		if _, err := file.Write(key); err != nil {
			return "", nil, err
		}

		// write value length
		valLen := uint32(len(value))
		if err := binary.Write(file, binary.LittleEndian, valLen); err != nil {
			return "", nil, err
		}

		// write value
		if _, err := file.Write(value); err != nil {
			return "", nil, err
		}
	}

	return filename, filter, nil
}
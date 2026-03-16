package compaction

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type Entry struct {
	key   []byte
	value []byte
}

func readAllEntries(filename string) ([]Entry, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry

	for {
		var keyLen uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(file, key); err != nil {
			return nil, err
		}

		var valLen uint32
		if err := binary.Read(file, binary.LittleEndian, &valLen); err != nil {
			return nil, err
		}

		value := make([]byte, valLen)
		if _, err := io.ReadFull(file, value); err != nil {
			return nil, err
		}

		entries = append(entries, Entry{key: key, value: value})
	}

	return entries, nil
}

func writeEntries(entries []Entry) (string, error) {

	filename := fmt.Sprintf("%d.sst", time.Now().UnixNano())

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for _, entry := range entries {

		keyLen := uint32(len(entry.key))
		if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
			return "", err
		}

		if _, err := file.Write(entry.key); err != nil {
			return "", err
		}

		valLen := uint32(len(entry.value))
		if err := binary.Write(file, binary.LittleEndian, valLen); err != nil {
			return "", err
		}

		if _, err := file.Write(entry.value); err != nil {
			return "", err
		}
	}

	// force flush to disk before returning
	if err := file.Sync(); err != nil {
		return "", err
	}

	return filename, nil
}

func Compact(sstables []string) (string, error) {

	fmt.Println("Starting compaction of", len(sstables), "SSTables...")

	merged := make(map[string]Entry)
	seen := make(map[string]bool)

	for _, filename := range sstables {

		entries, err := readAllEntries(filename)
		if err != nil {
			return "", err
		}

		for _, entry := range entries {
			key := string(entry.key)
			if !seen[key] {
				seen[key] = true
				if entry.value != nil {
					merged[key] = entry
				}
			}
		}
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []Entry
	for _, k := range keys {
		result = append(result, merged[k])
	}

	newFile, err := writeEntries(result)
	if err != nil {
		return "", err
	}

	for _, filename := range sstables {
		if err := os.Remove(filename); err != nil {
			return "", err
		}
	}

	fmt.Println("Compaction done. New SSTable:", newFile)

	return newFile, nil
}
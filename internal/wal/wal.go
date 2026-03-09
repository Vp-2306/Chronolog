package wal

import (
	"bufio"
	"os"
	"strings"

	"github.com/Vp-2306/Chronolog/internal/memtable"
)

type WAL struct {
	file   *os.File
	writer *bufio.Writer
}

func NewWAL(path string) (*WAL, error) {

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)

	return &WAL{
		file:   file,
		writer: writer,
	}, nil
}

func (w *WAL) Append(record []byte) error {

	_, err := w.writer.Write(record)

	if err != nil {
		return err
	}

	return w.writer.Flush()
}

func (w *WAL) Close() error {
	return w.file.Close()
}

func (w *WAL) Recover(mem *memtable.MemTable) error {

	file, err := os.Open(w.file.Name())
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := scanner.Text()

		parts := strings.Split(line, " ")

		if len(parts) < 3 {
			continue
		}

		command := parts[0]
		key := parts[1]
		value := parts[2]

		if command == "PUT" {
			mem.Put([]byte(key), []byte(value))
		}
	}

	return nil
}
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
	if err := w.writer.Flush(); err != nil{
		return err
	}
	return w.file.Close()
}

func (w *WAL) Truncate() error {
    if err := w.writer.Flush(); err != nil {
        return err
    }
    if err := w.file.Close(); err != nil {
        return err
    }
    if err := os.Truncate(w.file.Name(), 0); err != nil {
        return err
    }
    // reopen so WAL is still usable after truncation
    file, err := os.OpenFile(
        w.file.Name(),
        os.O_CREATE|os.O_APPEND|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return err
    }
    w.file = file
    w.writer = bufio.NewWriter(file)
    return nil
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

		if len(parts) < 2 {
			continue
		}

		command := parts[0]
		key := parts[1]

		if command == "PUT" && len(parts) >= 3 {
			value := parts[2]
			mem.Put([]byte(key), []byte(value))
		} else if command == "DELETE" {
			mem.Delete([]byte(key))
		}
	}

	return nil
}
package chronolog

import (
	"fmt"
	"os"
	"sort"

	"github.com/Vp-2306/Chronolog/internal/memtable"
	"github.com/Vp-2306/Chronolog/internal/sstable"
	"github.com/Vp-2306/Chronolog/internal/wal"
)

type Engine struct {
	mem *memtable.MemTable
	wal *wal.WAL
	sstables []string
}

func NewEngine(walPath string) (*Engine, error) {

	mem := memtable.NewMemTable()

	sstables, err := loadSSTables()
	if err != nil {
		return nil, err
	}

	// recover from WAL if it exists
	if _, err := os.Stat(walPath); err == nil {

		fmt.Println("Recovering from WAL...")

		w, err := wal.NewWAL(walPath)
		if err != nil {
			return nil, err
		}

		if err := w.Recover(mem); err != nil {
			return nil, err
		}

		if err := w.Truncate(); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}
	}

	w, err := wal.NewWAL(walPath)
	if err != nil {
		return nil, err
	}

	return &Engine{
		mem: mem,
		wal: w,
		sstables: sstables,
	}, nil
}

func loadSSTables() ([]string, error) {

	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 &&
			entry.Name()[len(entry.Name())-4:] == ".sst" {
			files = append(files, entry.Name())
		}
	}

	// sort newest first
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	return files, nil
}

func (e *Engine) Put(key []byte, value []byte) error {

	// write to WAL first
	record := fmt.Sprintf("PUT %s %s\n", key, value)
	if err := e.wal.Append([]byte(record)); err != nil {
		return err
	}

	// write to memtable
	e.mem.Put(key, value)

	// check if memtable needs flushing
	if e.mem.ShouldFlush() {
		if err := e.flush(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) Get(key []byte) ([]byte, error) {

	// check memtable first
	val := e.mem.Get(key)
	if val != nil {
		return val, nil
	}

	// search SSTables newest first
	for _, filename := range e.sstables {
		value, found, err := sstable.ReadSSTable(filename, key)
		if err != nil {
			return nil, err
		}
		if found {
			// nil value means tombstone — key was deleted
			if value == nil {
				return nil, fmt.Errorf("key not found")
			}
			return value, nil
		}
	}

	return nil, fmt.Errorf("key not found")
}


func (e *Engine) Delete(key []byte) error {

	record := fmt.Sprintf("DELETE %s\n", key)
	if err := e.wal.Append([]byte(record)); err != nil {
		return err
	}

	e.mem.Delete(key)

	return nil
}

func (e *Engine) flush() error {

	fmt.Println("Flushing memtable to SSTable...")

	filename, err := sstable.WriteSSTable(e.mem)
	if err != nil {
		return err
	}

	fmt.Println("SSTable written:", filename)

	// add new SSTable to front of list — newest first
	e.sstables = append([]string{filename}, e.sstables...)

	// reset memtable
	e.mem = memtable.NewMemTable()

	// truncate WAL
	if err := e.wal.Truncate(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) Close() error {
	if e.mem.HasData() && e.mem.ShouldFlush() {
		if err := e.flush(); err != nil {
			return err
		}
	}
	return e.wal.Close()
}
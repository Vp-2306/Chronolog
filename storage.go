package chronolog

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/Vp-2306/Chronolog/internal/bloom"
	"github.com/Vp-2306/Chronolog/internal/compaction"
	"github.com/Vp-2306/Chronolog/internal/memtable"
	"github.com/Vp-2306/Chronolog/internal/metrics"
	"github.com/Vp-2306/Chronolog/internal/sstable"
	"github.com/Vp-2306/Chronolog/internal/wal"
)

const maxLevel0Files = 4

type Engine struct {
	mem      *memtable.MemTable
	wal      *wal.WAL
	sstables []string
	filters  map[string]*bloom.BloomFilter
}

func NewEngine(walPath string) (*Engine, error) {

	mem := memtable.NewMemTable()

	sstables, err := loadSSTables()
	if err != nil {
		return nil, err
	}

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
		mem:      mem,
		wal:      w,
		sstables: sstables,
		filters:  make(map[string]*bloom.BloomFilter),
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

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	return files, nil
}

func (e *Engine) Put(key []byte, value []byte) error {

	start := time.Now()

	record := fmt.Sprintf("PUT %s %s\n", key, value)
	if err := e.wal.Append([]byte(record)); err != nil {
		return err
	}

	e.mem.Put(key, value)

	// update memtable size metric
	metrics.Global.UpdateMemTableSize(int64(e.mem.Size()))

	if e.mem.ShouldFlush() {
		if err := e.flush(); err != nil {
			return err
		}
	}

	// record write latency
	metrics.Global.RecordWrite(time.Since(start))

	return nil
}

func (e *Engine) Get(key []byte) ([]byte, error) {

	start := time.Now()

	val := e.mem.Get(key)
	if val != nil {
		metrics.Global.RecordRead(time.Since(start))
		return val, nil
	}

	for _, filename := range e.sstables {

		if filter, ok := e.filters[filename]; ok {
			if !filter.MightContain(key) {
				metrics.Global.RecordBloomHit()
				continue
			}
		}

		metrics.Global.RecordBloomMiss()

		value, found, err := sstable.ReadSSTable(filename, key)
		if err != nil {
			return nil, err
		}
		if found {
			metrics.Global.RecordRead(time.Since(start))
			if value == nil {
				return nil, fmt.Errorf("key not found")
			}
			return value, nil
		}
	}

	metrics.Global.RecordRead(time.Since(start))
	return nil, fmt.Errorf("key not found")
}

func (e *Engine) Delete(key []byte) error {

	record := fmt.Sprintf("DELETE %s\n", key)
	if err := e.wal.Append([]byte(record)); err != nil {
		return err
	}

	e.mem.Delete(key)
	metrics.Global.RecordDelete()

	return nil
}

func (e *Engine) flush() error {

	fmt.Println("Flushing memtable to SSTable...")

	filename, filter, err := sstable.WriteSSTable(e.mem)
	if err != nil {
		return err
	}

	fmt.Println("SSTable written:", filename)

	e.filters[filename] = filter
	e.sstables = append([]string{filename}, e.sstables...)
	e.mem = memtable.NewMemTable()

	metrics.Global.RecordFlush()
	metrics.Global.UpdateMemTableSize(0)

	if err := e.wal.Truncate(); err != nil {
		return err
	}

	if len(e.sstables) >= maxLevel0Files {
		if err := e.compact(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) compact() error {

	filesRemoved := int64(len(e.sstables))

	newFile, err := compaction.Compact(e.sstables)
	if err != nil {
		return err
	}

	for _, filename := range e.sstables {
		delete(e.filters, filename)
	}

	e.sstables = []string{newFile}

	newFilter, err := buildFilter(newFile)
	if err != nil {
		return err
	}

	e.filters[newFile] = newFilter

	metrics.Global.RecordCompaction(filesRemoved)

	return nil
}

func buildFilter(filename string) (*bloom.BloomFilter, error) {

	filter := bloom.NewBloomFilter()

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

		filter.Add(key)

		var valLen uint32
		if err := binary.Read(file, binary.LittleEndian, &valLen); err != nil {
			return nil, err
		}

		value := make([]byte, valLen)
		if _, err := io.ReadFull(file, value); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func (e *Engine) Close() error {
	if e.mem.HasData() && e.mem.ShouldFlush() {
		if err := e.flush(); err != nil {
			return err
		}
	}
	return e.wal.Close()
}
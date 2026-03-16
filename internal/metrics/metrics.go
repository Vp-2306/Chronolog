package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// operation counts
	TotalWrites    int64
	TotalReads     int64
	TotalDeletes   int64
	TotalFlushes   int64
	TotalCompactions int64

	// bloom filter stats
	BloomFilterHits  int64 // skipped a file correctly
	BloomFilterMisses int64 // had to open a file

	// latency tracking in nanoseconds
	TotalWriteLatency int64
	TotalReadLatency  int64

	// sstable stats
	SSTables int64

	// memtable size
	MemTableSize int64
}

var Global = &Metrics{}

func (m *Metrics) RecordWrite(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalWrites++
	m.TotalWriteLatency += duration.Nanoseconds()
}

func (m *Metrics) RecordRead(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalReads++
	m.TotalReadLatency += duration.Nanoseconds()
}

func (m *Metrics) RecordDelete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalDeletes++
}

func (m *Metrics) RecordFlush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalFlushes++
	m.SSTables++
}

func (m *Metrics) RecordCompaction(filesRemoved int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalCompactions++
	// after compaction many files become 1
	m.SSTables = m.SSTables - filesRemoved + 1
}

func (m *Metrics) RecordBloomHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BloomFilterHits++
}

func (m *Metrics) RecordBloomMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BloomFilterMisses++
}

func (m *Metrics) UpdateMemTableSize(size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MemTableSize = size
}

func (m *Metrics) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgWriteLatency := int64(0)
	if m.TotalWrites > 0 {
		avgWriteLatency = m.TotalWriteLatency / m.TotalWrites
	}

	avgReadLatency := int64(0)
	if m.TotalReads > 0 {
		avgReadLatency = m.TotalReadLatency / m.TotalReads
	}

	bloomHitRate := float64(0)
	totalBloom := m.BloomFilterHits + m.BloomFilterMisses
	if totalBloom > 0 {
		bloomHitRate = float64(m.BloomFilterHits) / float64(totalBloom) * 100
	}

	return map[string]interface{}{
		"total_writes":         m.TotalWrites,
		"total_reads":          m.TotalReads,
		"total_deletes":        m.TotalDeletes,
		"total_flushes":        m.TotalFlushes,
		"total_compactions":    m.TotalCompactions,
		"bloom_filter_hits":    m.BloomFilterHits,
		"bloom_filter_misses":  m.BloomFilterMisses,
		"bloom_hit_rate":       bloomHitRate,
		"avg_write_latency_ns": avgWriteLatency,
		"avg_read_latency_ns":  avgReadLatency,
		"sstable_count":        m.SSTables,
		"memtable_size_bytes":  m.MemTableSize,
	}
}
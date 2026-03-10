package memtable

import "sync"

const MaxSize = 4 * 1024 * 1024 // 4MB

type MemTable struct{
	list *SkipList
	mu sync.RWMutex
	size int 
}

func  NewMemTable() *MemTable {

	return &MemTable{
		list: NewSkipList(),
	}
}

func (m *MemTable) Put(key []byte, value []byte){

	m.mu.Lock()
	defer m.mu.Unlock()

	m.list.Insert(key,value)
	m.size += len(key) + len(value)
}

func (m *MemTable) Get(key []byte) []byte{

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.list.Search(key)
}

func (m *MemTable) Delete(key []byte){
	m.mu.Lock()
	defer m.mu.Unlock()

	m.list.Insert(key,nil)
	m.size += len(key)
}

func (m *MemTable) ShouldFlush() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.size >= MaxSize
}

func (m *MemTable) NewIterator() *Iterator {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.list.NewIterator()
}

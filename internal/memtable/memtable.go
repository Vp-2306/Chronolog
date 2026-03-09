package memtable

import "sync"

type MemTable struct{
	list *SkipList
	mu sync.RWMutex
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
}

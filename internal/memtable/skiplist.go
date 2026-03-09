package memtable

import (
	"bytes"
	"math/rand"
)

const MaxLevel = 16
const Probability = 0.5

type Node struct {
	key   []byte
	value []byte

	forward []*Node
}

type SkipList struct {
	header *Node
	level  int
}

func NewSkipList() *SkipList {

	header := &Node{
		forward: make([]*Node, MaxLevel),
	}

	return &SkipList{
		header: header,
		level:  1,
	}
}

func randomLevel() int {

	level := 1

	for rand.Float64() < Probability && level < MaxLevel {
		level++
	}

	return level
}

func (s *SkipList) Search(key []byte) []byte {

	current := s.header

	for i := s.level - 1; i >= 0; i-- {

		for current.forward[i] != nil &&
			bytes.Compare(current.forward[i].key, key) < 0 {

			current = current.forward[i]
		}
	}

	current = current.forward[0]

	if current != nil && bytes.Equal(current.key, key) {
		return current.value
	}

	return nil
}

func (s *SkipList) Insert(key []byte, value []byte) {

	update := make([]*Node, MaxLevel)

	current := s.header

	for i := s.level - 1; i >= 0; i-- {

		for current.forward[i] != nil &&
			bytes.Compare(current.forward[i].key, key) < 0 {

			current = current.forward[i]
		}

		update[i] = current
	}

	current = current.forward[0]

	if current != nil && bytes.Equal(current.key, key) {
		current.value = value
		return
	}

	level := randomLevel()

	if level > s.level {

		for i := s.level; i < level; i++ {
			update[i] = s.header
		}

		s.level = level
	}

	newNode := &Node{
		key:     key,
		value:   value,
		forward: make([]*Node, level),
	}

	for i := 0; i < level; i++ {

		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
}


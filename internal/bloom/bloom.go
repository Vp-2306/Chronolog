	package bloom

	import (
		"hash/fnv"
	)

	const(
		arraySize = 1024
		numHashes = 3
	)

	type BloomFilter struct {
		bits []bool
	}

	func NewBloomFilter() *BloomFilter {
		return &BloomFilter{
			bits: make([]bool, arraySize),
		}
	}

	func getPositions(key []byte) []uint {

		positions := make([]uint, numHashes)

		h1 := fnv.New64a()
		h1.Write(key)
		hash1 := h1.Sum64()

		h2 := fnv.New64a()
		h2.Write(key)
		h2.Write([]byte{42})
		hash2 := h2.Sum64()

		for i := uint64(0); i < numHashes; i++ {
			positions[i] = uint((hash1 + i*hash2) % arraySize)
		}

		return positions
	}

	func (bf *BloomFilter) Add(key []byte){
		for _, pos := range getPositions(key){
			bf.bits[pos] = true
		}
	}

	func (bf *BloomFilter) MightContain(key []byte) bool {
		for _,pos := range getPositions(key){
			if !bf.bits[pos]{
				return false
			}
		}
		return true
	}
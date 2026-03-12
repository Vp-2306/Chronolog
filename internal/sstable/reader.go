package sstable

import(
	"encoding/binary"
	"io"
	"os"
)

func ReadSSTable(filename string, searchKey []byte) ([]byte, bool, error){

	file, err := os.Open(filename)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	for{
		var keyLen uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
			if err == io.EOF{
				break
			}
			return nil, false, err 
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(file, key); err != nil{
			return nil, false, err
		}
		var valLen uint32
		if err := binary.Read(file, binary.LittleEndian, &valLen); err != nil {
			return nil, false, err
		}

		value := make([]byte, valLen)
		if _, err := io.ReadFull(file, value); err != nil {
			return nil, false, err
		}

		if string(key) == string(searchKey) {
			return value, true, nil
		}
	}

	return nil, false, nil

}
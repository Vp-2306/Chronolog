package chronolog

type StorageEngine interface {

	Put(key []byte, value []byte) error

	Get(key []byte) ([]byte, error)

	Delete(key []byte) error
}
package db

type KvStore interface {
	Search(key string) ([]string, error)
	Set(key string, value interface{}) error
	Get(key string) (string, error)
	TotalKeys() int
}

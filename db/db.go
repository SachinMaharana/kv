package db

import (
	"log"

	"github.com/go-redis/redis"
)

type Repository interface {
	Search(key string) ([]string, error)
	Set(key string, value interface{}) error
	Get(key string) (string, error)
	Total() int
}

// db backed by redis
type repository struct {
	Client redis.Cmdable
}

func (r *repository) Total() int {
	keys := []string{}
	iter := r.Client.Scan(0, "*", 0).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Println("Error in getting total keys metrics")
	}
	return len(keys)
}

func NewRedisRepository(Client redis.Cmdable) Repository {
	return &repository{Client}
}

func (r *repository) Set(key string, value interface{}) error {
	log.Println("SET", key)

	return r.Client.Set(key, value, 0).Err()
}

func (r *repository) Search(key string) ([]string, error) {
	var cursor uint64
	var n int
	var keys []string
	var err error

	for {
		keys, cursor, err = r.Client.Scan(cursor, key, 10).Result()
		n += len(keys)
		if cursor == 0 {
			break
		}
	}
	return keys, err
}

// Get attaches the redis repository and get the data
func (r *repository) Get(key string) (string, error) {
	get := r.Client.Get(key)
	return get.Result()
}

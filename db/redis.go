package db

import (
	"log"

	"github.com/go-redis/redis"
)

// db backed by redis
type redisStore struct {
	Client redis.Cmdable
}

func NewRedisStore(Client redis.Cmdable) KvStore {
	return &redisStore{Client}
}

// TotalKeys returns total number of keys in db
func (r *redisStore) TotalKeys() int {
	keys := []string{}
	iter := r.Client.Scan(0, "*", 0).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		//TODO: how to handle this?
		log.Println(err.Error())
	}
	return len(keys)
}

func (r *redisStore) Set(key string, value interface{}) error {
	return r.Client.Set(key, value, 0).Err()
}

// Get attaches the redis store and get the data
func (r *redisStore) Get(key string) (string, error) {
	get := r.Client.Get(key)
	return get.Result()
}

func (r *redisStore) Search(key string) ([]string, error) {
	var keys []string
	var err error
	iter := r.Client.Scan(0, key, 0).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}
	if err = iter.Err(); err != nil {
		log.Println(err.Error())
		return []string{}, err
	}
	return keys, nil
}

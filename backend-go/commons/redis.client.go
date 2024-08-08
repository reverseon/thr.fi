package commons

import (
	"os"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
)

var lock = &sync.Mutex{}

var instance *redis.Client

func GetRedisClient() *redis.Client {
	if instance == nil {
		db_index, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			panic(err)
		}
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = redis.NewClient(&redis.Options{
				Addr:     os.Getenv("REDIS_ADDR"),
				Username: os.Getenv("REDIS_USERNAME"),
				Password: os.Getenv("REDIS_PASSWORD"),
				DB:       db_index,
			})
		}
	}
	return instance
}

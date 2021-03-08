package connections

import (
	"os"

	"github.com/go-redis/redis/v8"
)

//Redis ...
func Redis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	return rdb
}

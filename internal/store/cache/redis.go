package cache

import "github.com/go-redis/redis/v8"

func NewCacheClient(addr, pwd string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
}

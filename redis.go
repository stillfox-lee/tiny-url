package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8/v2"
)

func InitRedisClient(host, port, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", host, port),
		Password:    password,
		DB:          db,
		DialTimeout: 3 * time.Second,
	})
	rdb.AddHook(apmgoredis.NewHook())
	result := rdb.Ping(context.TODO())
	// ensure key "range:100000-1000000" exist
	cmd := rdb.Set(context.Background(), "range:100000-1000000", 100000, 0)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return rdb, result.Err()
}

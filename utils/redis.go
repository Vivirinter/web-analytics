package utils

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

var RC redis.Conn

var KS = ":"

func GetRedisConnection() redis.Conn {
	rc, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v\n", err)
	}
	return rc
}

package core

import (
	"log"

	"github.com/garyburd/redigo/redis"
)

// SetStatus writes server status in redis cache
func SetStatus(redisURL, status, serverStateKey string) {
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		log.Printf("Redis connection error: %s", err)
		return
	}
	defer conn.Close()
	_, err = conn.Do("HSET", serverStateKey, "status", status)
	if err != nil {
		log.Printf("Set status error: %s", err)
	}
}

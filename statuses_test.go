package core

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

func TestSetStatus(t *testing.T) {
	redisURL := "redis://localhost:6379/0"
	serverStateKey := "server_state_hashid"
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		t.Errorf("Redis connection error: %s", err)
	}
	defer conn.Close()
	status := "Test"
	SetStatus(redisURL, status, serverStateKey)
	resp, err := redis.String(conn.Do("HGET", serverStateKey, "status"))
	if err != nil {
		t.Error("Error retrieving status")
	}
	if resp != status {
		t.Errorf("Wrong status.\nExpected: %s\nActual: %s", status, resp)
	}
}

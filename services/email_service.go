package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var RedisCtx = context.Background()

// ConnectRedis initializes the Redis client.
// If Redis is not available, the app continues without caching.
func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(RedisCtx, 3*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("⚠️  Redis not available, caching disabled:", err)
		RedisClient = nil
	} else {
		fmt.Println("✅ Connected to Redis successfully!")
	}
}

// CacheSet stores a value in Redis with a TTL
func CacheSet(key string, value string, ttl time.Duration) error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Set(RedisCtx, key, value, ttl).Err()
}

// CacheGet retrieves a value from Redis
func CacheGet(key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis not available")
	}
	return RedisClient.Get(RedisCtx, key).Result()
}

// CacheDelete removes a key from Redis
func CacheDelete(key string) error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Del(RedisCtx, key).Err()
}

// CacheDeletePattern removes all keys matching a pattern
func CacheDeletePattern(pattern string) error {
	if RedisClient == nil {
		return nil
	}
	keys, err := RedisClient.Keys(RedisCtx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return RedisClient.Del(RedisCtx, keys...).Err()
	}
	return nil
}

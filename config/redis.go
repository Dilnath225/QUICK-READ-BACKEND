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
	if os.Getenv("USE_REDIS") == "false" {
		fmt.Println("ℹ️  Redis disabled by environment variable")
		return
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379" // Use 127.0.0.1 instead of localhost to avoid IPv6 [::1] mismatch
	}

	password := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
		MaxRetries: 0, // Disable retries for initial connection check to avoid log clutter
	})

	ctx, cancel := context.WithTimeout(RedisCtx, 3*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("⚠️  Redis not available at %s, caching disabled\n", addr)
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

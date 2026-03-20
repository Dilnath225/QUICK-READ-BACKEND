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

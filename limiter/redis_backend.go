package limiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisBackend struct {
	client *redis.Client
}

func NewRedisBackend(redisURL string) Backend {
	client := redis.NewClient(&redis.Options{Addr: redisURL})
	return &redisBackend{client: client}
}

func (r *redisBackend) Take(key string, tokens int) (bool, error) {
	ctx := context.Background()

	script := `
		local tokens = redis.call("GET", KEYS[1])
		if not tokens then
			tokens = tonumber(ARGV[2])
			redis.call("SET", KEYS[1], tokens, "EX", ARGV[3])
		end
		if tonumber(tokens) >= tonumber(ARGV[1]) then
			redis.call("DECRBY", KEYS[1], ARGV[1])
			return 1
		else
			return 0
		end
	`
	res, err := r.client.Eval(ctx, script, []string{key}, tokens, tokens, int(time.Minute.Seconds())).Int()
	return res == 1, err
}

func (r *redisBackend) Reset(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

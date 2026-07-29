package redkit

import (
	"context"
	"fmt"

	"github.com/noble-gase/neon/helper"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Discard 丢弃数据，不缓存
const Discard = helper.NilError("redkit: discarded")

var sf singleflight.Group

var script = redis.NewScript(`
redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
if redis.call('TTL', KEYS[1]) == -1 then
    redis.call('EXPIRE', KEYS[1], ARGV[3])
end
`)

func doSF[T any](ctx context.Context, key string, fn func() (any, error)) (T, error) {
	var ret T

	select {
	case <-ctx.Done():
		return ret, ctx.Err()
	case result := <-sf.DoChan(key, fn):
		if result.Err != nil {
			return ret, result.Err
		}

		data, ok := result.Val.(T)
		if !ok {
			return ret, fmt.Errorf("redkit: unexpected result type %T", result.Val)
		}
		return data, nil
	}
}

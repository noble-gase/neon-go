package redkit

import (
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

func doSF[T any](key string, fn func() (any, error)) (T, error) {
	var ret T

	value, err, _ := sf.Do(key, fn)
	if err != nil {
		return ret, err
	}

	data, ok := value.(T)
	if !ok {
		return ret, fmt.Errorf("redkit: unexpected result type %T", value)
	}
	return data, nil
}

package redkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get 读取缓存，未命中时通过 loader 回源并写入缓存。
// 回源使用 singleflight 去重，调用方需保证 key 全局唯一（跨客户端、跨操作、跨类型），否则可能共享到非预期结果
func Get[T any](ctx context.Context, uc redis.UniversalClient, key string, loader func(ctx context.Context) (T, error), ttl time.Duration) (T, error) {
	var ret T

	str, err := uc.Get(ctx, key).Result()
	if err == nil {
		if _err := json.Unmarshal([]byte(str), &ret); _err != nil {
			return ret, fmt.Errorf("unmarshal(%s): %w", str, _err)
		}
		return ret, nil
	}
	if !errors.Is(err, redis.Nil) {
		return ret, err
	}

	// 缓存未命中
	return doSF[T](key, func() (any, error) {
		// 调用 loader 回源数据
		data, _err := loader(ctx)
		if _err != nil {
			if errors.Is(_err, Discard) {
				return data, nil
			}
			return nil, _err
		}

		// 缓存数据
		b, _err := json.Marshal(data)
		if _err != nil {
			return nil, fmt.Errorf("marshal(%+v): %w", data, _err)
		}
		if _err = uc.Set(ctx, key, string(b), ttl).Err(); _err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "[redkit:Get] set data failed", slog.String("key", key), slog.String("value", string(b)), slog.Any("error", _err))
		}
		return data, nil
	})
}

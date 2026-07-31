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

// HGet 读取 Hash 字段缓存，未命中时通过 loader 回源并写入缓存。
// 回源使用 singleflight 去重，调用方需保证 key:field 全局唯一（跨客户端、跨操作、跨类型），否则可能共享到非预期结果
func HGet[T any](ctx context.Context, uc redis.UniversalClient, key, field string, loader func(ctx context.Context) (T, error), ttl time.Duration) (T, error) {
	var ret T

	str, err := uc.HGet(ctx, key, field).Result()
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
	sfKey := key + ":" + field
	return doSF[T](sfKey, func() (any, error) {
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

		if ttl > 0 {
			sec := int64(ttl.Seconds())
			if sec <= 0 {
				sec = 1
			}
			_err = script.Run(ctx, uc, []string{key}, field, string(b), sec).Err()
		} else {
			_err = uc.HSet(ctx, key, field, string(b)).Err()
		}
		if _err != nil && !errors.Is(_err, redis.Nil) {
			slog.LogAttrs(ctx, slog.LevelError, "[redkit:HGet] hset data failed", slog.String("key", key), slog.String("field", field), slog.String("value", string(b)), slog.Any("error", _err))
		}

		return data, nil
	})
}

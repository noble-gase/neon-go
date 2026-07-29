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

func Get[T any](ctx context.Context, uc redis.UniversalClient, key string, fn func(ctx context.Context) (T, error), ttl time.Duration) (T, error) {
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
	return doSF[T](ctx, key, func() (any, error) {
		// 调用fn获取数据
		data, _err := fn(ctx)
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

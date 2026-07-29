package redkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func MGetMap[T any](ctx context.Context, uc redis.UniversalClient, keys []string) (map[string]T, error) {
	values, err := uc.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	if len(values) != len(keys) {
		return nil, errors.New("the number of keys and values mismatch")
	}

	ret := make(map[string]T, len(keys))
	for i, k := range keys {
		if v := values[i]; v != nil {
			if s, ok := v.(string); ok && len(s) != 0 {
				var val T
				if err = json.Unmarshal([]byte(s), &val); err != nil {
					return nil, fmt.Errorf("unmarshal(%s): %w", s, err)
				}
				ret[k] = val
			}
		}
	}
	return ret, nil
}

func MGetStrMap(ctx context.Context, uc redis.UniversalClient, keys []string) (map[string]string, error) {
	values, err := uc.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	if len(values) != len(keys) {
		return nil, errors.New("the number of keys and values mismatch")
	}

	ret := make(map[string]string, len(keys))
	for i, k := range keys {
		if v := values[i]; v != nil {
			if s, ok := v.(string); ok && len(s) != 0 {
				ret[k] = s
			}
		}
	}
	return ret, nil
}

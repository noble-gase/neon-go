package redkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func HGetAll[T any](ctx context.Context, uc redis.UniversalClient, key string) (map[string]T, error) {
	data, err := uc.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]T, len(data))
	for k, v := range data {
		var val T
		if err = json.Unmarshal([]byte(v), &val); err != nil {
			return nil, fmt.Errorf("unmarshal(%s): %w", v, err)
		}
		ret[k] = val
	}
	return ret, nil
}

func HMGetMap[T any](ctx context.Context, uc redis.UniversalClient, key string, fields []string) (map[string]T, error) {
	values, err := uc.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	if len(values) != len(fields) {
		return nil, errors.New("the number of fields and values mismatch")
	}

	ret := make(map[string]T, len(fields))
	for i, k := range fields {
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

func HMGetStrMap(ctx context.Context, uc redis.UniversalClient, key string, fields []string) (map[string]string, error) {
	values, err := uc.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	if len(values) != len(fields) {
		return nil, errors.New("the number of fields and values mismatch")
	}

	ret := make(map[string]string, len(fields))
	for i, k := range fields {
		if v := values[i]; v != nil {
			if s, ok := v.(string); ok && len(s) != 0 {
				ret[k] = s
			}
		}
	}
	return ret, nil
}

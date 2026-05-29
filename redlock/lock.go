package redlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/noble-gase/neon/helper"
	"github.com/redis/go-redis/v9"
)

// Nil 未获取到锁
var Nil = helper.NilError("redlock: nil")

var script = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

// RedLock 基于「Redis」实现的分布式锁
//
// 注意：单个 RedLock 实例不是并发安全的，同一实例不应被多个 goroutine 共享
type RedLock struct {
	uc    redis.UniversalClient
	key   string
	ttl   time.Duration
	token string
}

func (l *RedLock) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done(): // timeout or canceled
		return context.Cause(ctx)
	default:
	}

	if err := l.setnx(ctx); err != nil {
		return err
	}
	if len(l.token) != 0 {
		return nil
	}
	return Nil
}

func (l *RedLock) TryAcquire(ctx context.Context, attempts int, duration time.Duration) error {
	threshold := attempts - 1
	for i := range attempts {
		select {
		case <-ctx.Done(): // timeout or canceled
			return context.Cause(ctx)
		default:
		}

		// attempt to acquire lock
		if err := l.setnx(ctx); err != nil {
			return err
		}
		if len(l.token) != 0 {
			return nil
		}
		if i < threshold {
			time.Sleep(duration)
		}
	}
	return Nil
}

func (l *RedLock) Release(ctx context.Context) error {
	if len(l.token) == 0 {
		return nil
	}

	defer func() {
		l.token = "" // clear token
	}()
	return script.Run(context.WithoutCancel(ctx), l.uc, []string{l.key}, l.token).Err()
}

func (l *RedLock) setnx(ctx context.Context) error {
	l.token = "" // clear token

	token := uuid.New().String()

	_, err := l.uc.SetArgs(ctx, l.key, token, redis.SetArgs{
		Mode: "NX",
		TTL:  l.ttl,
	}).Result()

	// 设置成功
	if err == nil {
		l.token = token
		return nil
	}

	// key已存在，未设置
	if errors.Is(err, redis.Nil) {
		return nil
	}

	// 异常错误，尝试GET一次：避免因网络错误导致误加锁
	val, getErr := l.uc.Get(ctx, l.key).Result()
	if getErr != nil {
		if errors.Is(getErr, redis.Nil) {
			return err
		}
		return fmt.Errorf("SET-NX: %w; GET: %w", err, getErr)
	}
	// NOTE: if SET timed out but actually succeeded, the remaining TTL
	// will be less than l.ttl due to network round-trip delay.
	if val == token {
		l.token = token
	}
	return nil
}

// New 返回一个Redis分布式锁
func New(uc redis.UniversalClient, key string, ttl time.Duration) *RedLock {
	mutex := &RedLock{
		uc:  uc,
		key: key,
		ttl: ttl,
	}
	if mutex.ttl <= 0 {
		mutex.ttl = time.Second * 10
	}
	return mutex
}

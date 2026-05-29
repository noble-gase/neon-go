package redkit

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/noble-gase/neon/helper"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"golang.org/x/sync/singleflight"
)

var sf singleflight.Group

// Discard 丢弃数据，不缓存
const Discard = helper.NilError("redkit: discarded")

var script = redis.NewScript(`
redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
if redis.call('TTL', KEYS[1]) == -1 then
    redis.call('EXPIRE', KEYS[1], ARGV[3])
end
`)

type Config struct {
	// Addrs 地址
	Addrs []string `json:"addrs" mapstructure:"addrs"`
	// Options 选项
	Options Options `json:"options" mapstructure:"options"`
}

type Options struct {
	// DB 数据库
	DB int `json:"db" mapstructure:"db"`
	// Username 用户名
	Username string `json:"username" mapstructure:"username"`
	// Password 密码
	Password string `json:"password" mapstructure:"password"`
	// DialTimeout 连接超时时间（单位：秒）
	DialTimeout int `json:"dial_timeout" mapstructure:"dial_timeout"`
	// ReadTimeout 读取超时时间（单位：秒）
	ReadTimeout int `json:"read_timeout" mapstructure:"read_timeout"`
	// WriteTimeout 写入超时时间（单位：秒）
	WriteTimeout int `json:"write_timeout" mapstructure:"write_timeout"`
	// PoolSize 连接池大小
	PoolSize int `json:"pool_size" mapstructure:"pool_size"`
	// PoolTimeout 连接池超时时间（单位：秒）
	PoolTimeout int `json:"pool_timeout" mapstructure:"pool_timeout"`
	// MinIdleConns 最小空闲连接数
	MinIdleConns int `json:"min_idle_conns" mapstructure:"min_idle_conns"`
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	// MaxActiveConns 最大活跃连接数
	MaxActiveConns int `json:"max_active_conns" mapstructure:"max_active_conns"`
	// ConnMaxIdleTime 连接最大闲置时间（单位：秒）
	ConnMaxIdleTime int `json:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	// ConnMaxLifetime 连接最大生命时长（单位：秒）
	ConnMaxLifetime int `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	// InsecureSkipVerify 是否跳过证书验证
	InsecureSkipVerify bool `json:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
}

func NewClient(cfg *Config) (redis.UniversalClient, error) {
	opts := &redis.UniversalOptions{
		Addrs:           cfg.Addrs,
		DB:              cfg.Options.DB,
		Username:        cfg.Options.Username,
		Password:        cfg.Options.Password,
		DialTimeout:     time.Duration(cfg.Options.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.Options.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.Options.WriteTimeout) * time.Second,
		PoolSize:        cfg.Options.PoolSize,
		PoolTimeout:     time.Duration(cfg.Options.PoolTimeout) * time.Second,
		MinIdleConns:    cfg.Options.MinIdleConns,
		MaxIdleConns:    cfg.Options.MaxIdleConns,
		MaxActiveConns:  cfg.Options.MaxActiveConns,
		ConnMaxIdleTime: time.Duration(cfg.Options.ConnMaxIdleTime) * time.Second,
		ConnMaxLifetime: time.Duration(cfg.Options.ConnMaxLifetime) * time.Second,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	}
	if cfg.Options.InsecureSkipVerify {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	client := redis.NewUniversalClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

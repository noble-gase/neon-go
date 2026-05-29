package sqlkit

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/noble-gase/neon/sqlkit/internal"
)

// Config 数据库初始化配置
type Config struct {
	// Driver 驱动名称
	//  [-MySQL] mysql
	//  [-PgSQL] pgx
	//  [SQLite] sqlite3
	Driver string `json:"driver" mapstructure:"driver"`
	// DSN 数据源名称
	//
	//  [-MySQL] <username>:<password>@tcp(<host>:3306)/<db>?timeout=10s&charset=utf8mb4&parseTime=True&loc=Local
	//  [-PgSQL] postgres://<username>:<password>@<host>:5432/<dbname>
	//  [SQLite] file::memory:?cache=shared || file:</path/test.db>
	DSN string `json:"dsn" mapstructure:"dsn"`
	// Options 数据库连接池选项
	Options Options `json:"options" mapstructure:"options"`
}

type Options struct {
	// MaxOpenConns 设置最大可打开的连接数
	MaxOpenConns int `json:"max_open_conns" mapstructure:"max_open_conns"`
	// MaxIdleConns 连接池最大闲置连接数
	MaxIdleConns int `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	// ConnMaxIdleTime 连接最大闲置时间（单位：秒）
	ConnMaxIdleTime int `json:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	// ConnMaxLifetime 连接的最大生命时长（单位：秒）
	ConnMaxLifetime int `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

// NewDB returns a new sql.DB
func NewDB(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Options.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Options.MaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cfg.Options.ConnMaxIdleTime) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.Options.ConnMaxLifetime) * time.Second)

	return db, nil
}

func SetLogger(fn internal.LogFunc) {
	internal.Logger = fn
}

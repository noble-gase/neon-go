package sqlkit

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
)

type (
	DB = map[string]*sql.DB
	TX = map[string]*sql.Tx
)

// Transaction 执行数据库事务
func Transaction(ctx context.Context, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx) error, opts ...*sql.TxOptions) (err error) {
	var opt *sql.TxOptions
	if len(opts) != 0 {
		opt = opts[0]
	}

	tx, _err := db.BeginTx(ctx, opt)
	if _err != nil {
		err = fmt.Errorf("begin transaction: %w", _err)
		return
	}

	rollback := func(err error) error {
		if e := tx.Rollback(); e != nil {
			err = fmt.Errorf("%w; rollback: %w", err, e)
		}
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			// if panic, should rollback
			e := fmt.Errorf("transaction panic: %+v", r)
			err = fmt.Errorf("%w\n%s", rollback(e), string(debug.Stack()))
		}
	}()

	if e := fn(ctx, tx); e != nil {
		err = rollback(e)
		return
	}

	if e := tx.Commit(); e != nil {
		err = rollback(fmt.Errorf("commit: %w", e))
	}
	return
}

// TransactionX 执行多数据库事务
func TransactionX(ctx context.Context, db DB, fn func(ctx context.Context, tx TX) error, opts ...*sql.TxOptions) (err error) {
	var opt *sql.TxOptions
	if len(opts) != 0 {
		opt = opts[0]
	}

	tx := make(TX, len(db))
	for k, v := range db {
		if v == nil {
			err = fmt.Errorf("db(%s) is nil (forgotten initialize?)", k)
			return
		}

		x, e := v.BeginTx(ctx, opt)
		if e != nil {
			// 回滚已开启的事务
			for _, t := range tx {
				_ = t.Rollback()
			}
			err = fmt.Errorf("begin transaction (%s): %w", k, e)
			return
		}
		tx[k] = x
	}

	rollback := func(err error) error {
		for k, v := range tx {
			if e := v.Rollback(); e != nil {
				err = fmt.Errorf("%w; rollback(%s): %w", err, k, e)
			}
		}
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			// if panic, should rollback
			e := fmt.Errorf("transaction panic: %+v", r)
			err = fmt.Errorf("%w\n%s", rollback(e), string(debug.Stack()))
		}
	}()

	if e := fn(ctx, tx); e != nil {
		err = rollback(e)
		return
	}

	for k, v := range tx {
		if e := v.Commit(); e != nil {
			err = rollback(fmt.Errorf("commit(%s): %w", k, e))
			return
		}
	}
	return
}

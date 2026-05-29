package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/mysql"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/noble-gase/neon/sqlkit/internal"
)

// M 用于 mysql. 的 INSERT & UPDATE
type M map[Column]any

func (m M) Split() (cols ColumnList, vals []any) {
	cap := len(m)

	cols = make(ColumnList, 0, cap)
	vals = make([]any, 0, cap)

	for k, v := range m {
		cols = append(cols, k)
		vals = append(vals, v)
	}
	return
}

// Insert 插入记录
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 语句示例
//	table.Demo.INSERT(table.Demo.Name).VALUES("hello")
//	// or
//	table.Demo.INSERT(table.Demo.Name).MODEL(model.Demo{Name: "hello"})
//
//	// 批量插入
//	table.Demo.INSERT(table.Demo.Name).
//		VALUES("hello").
//		VALUES("world")
//	// or
//	table.Demo.INSERT(table.Demo.Name).MODELS([]model.Demo{
//		{Name: "hello"},
//		{Name: "world"},
//	})
//
//	// 执行方法
//	mysql.Insert(ctx, db, stmt)
func Insert(ctx context.Context, db qrm.DB, stmt InsertStatement) (int64, error) {
	var (
		ret sql.Result
		err error
	)

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	ret, err = stmt.ExecContext(ctx, db)
	if err != nil {
		return 0, err
	}

	id, _ := ret.LastInsertId()
	return id, nil
}

// Update 更新记录
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 语句示例
//	table.Demo.UPDATE(table.Demo.Name).SET("hello").WHERE(table.Demo.ID.EQ(jet.Int64(1)))
//	// or
//	table.Demo.UPDATE(table.Demo.Name).MODEL(model.Demo{Name: "hello"}).WHERE(table.Demo.ID.EQ(jet.Int64(1)))
//
//	// 执行方法
//	mysql.Update(ctx, db, stmt)
func Update(ctx context.Context, db qrm.DB, stmt UpdateStatement) (int64, error) {
	var (
		ret sql.Result
		err error
	)

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	ret, err = stmt.ExecContext(ctx, db)
	if err != nil {
		return 0, err
	}

	rows, _ := ret.RowsAffected()
	return rows, nil
}

// Delete 删除记录
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 语句示例
//	table.Demo.DELETE().WHERE(table.Demo.ID.EQ(jet.Int64(1)))
//
//	// 执行方法
//	mysql.Delete(ctx, db, stmt)
func Delete(ctx context.Context, db qrm.DB, stmt DeleteStatement) (int64, error) {
	var (
		ret sql.Result
		err error
	)

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	ret, err = stmt.ExecContext(ctx, db)
	if err != nil {
		return 0, err
	}

	rows, _ := ret.RowsAffected()
	return rows, nil
}

// FindOne 查询一条记录
//
// 注意：参数 T 必须为非指针类型
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 语句示例
//	table.Demo.SELECT(table.Demo.AllColumns).WHERE(table.Demo.ID.EQ(jet.Int64(1)))
//	// or
//	jet.SELECT(table.Demo.AllColumns).FROM(table.Demo).WHERE(table.Demo.ID.EQ(jet.Int64(1)))
//
//	// 执行方法
//	mysql.FindOne[model.Demo](ctx, db, stmt)
func FindOne[T any](ctx context.Context, db qrm.DB, stmt SelectStatement) (*T, error) {
	var (
		dest T
		err  error
	)

	stmt = stmt.LIMIT(1)

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	if err := stmt.QueryContext(ctx, db, &dest); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &dest, nil
}

// FindAll 查询多条记录
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 语句示例
//	table.Demo.SELECT(table.Demo.AllColumns).WHERE(table.Demo.Name.LIKE(jet.String("%hello%")))
//	// or
//	jet.SELECT(table.Demo.AllColumns).FROM(table.Demo).WHERE(table.Demo.Name.LIKE(jet.String("%hello%")))
//
//	// 执行方法
//	mysql.FindAll[*model.Demo](ctx, db, stmt)
func FindAll[T any](ctx context.Context, db qrm.DB, stmt SelectStatement) ([]T, error) {
	var (
		dest []T
		err  error
	)

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	if err := stmt.QueryContext(ctx, db, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// Count 返回记录数
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 执行方法
//	mysql.Count(ctx, db, func(count jet.SelectStatement) jet.SelectStatement {
//		return count.FROM(table.Demo.Table).WHERE(table.Demo.Name.LIKE(jet.String("%hello%")))
//	})
func Count(ctx context.Context, db qrm.DB, fn func(count SelectStatement) SelectStatement) (int64, error) {
	var (
		total struct {
			Count int64
		}
		err error
	)

	stmt := fn(SELECT(COUNT(STAR).AS("count")))

	start := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(stmt.DebugSql()), time.Since(start), err)
		}
	}()

	if err = stmt.QueryContext(ctx, db, &total); err != nil {
		return 0, err
	}
	return total.Count, nil
}

// Paginate 分页查询
//
//	// 导入模块
//	import (
//		jet "github.com/go-jet/jet/v2/mysql"
//		"github.com/noble-gase/neon/sqlkit/mysql"
//	)
//
//	// 执行方法
//	mysql.Paginate[*model.Demo](ctx, db, func(query jet.SelectStatement) jet.SelectStatement {
//		return query.FROM(table.Demo.Table).WHERE(table.Demo.Name.LIKE(jet.String("%hello%")))
//	}, page, size, table.Demo.AllColumns, table.Demo.ID.DESC())
func Paginate[T any](ctx context.Context, db qrm.DB, fn func(query SelectStatement) SelectStatement, page, size int, cols ColumnList, orderBy ...OrderByClause) ([]T, int64, error) {
	var (
		total struct {
			Count int64
		}
		countErr error
	)

	// 构建 count 查询
	countStmt := fn(SELECT(COUNT(STAR).AS("count")))

	countStart := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(countStmt.DebugSql()), time.Since(countStart), countErr)
		}
	}()

	countErr = countStmt.QueryContext(ctx, db, &total)
	if countErr != nil {
		return nil, 0, countErr
	}
	if total.Count == 0 {
		return []T{}, 0, nil
	}

	// 数据查询

	var (
		dest     []T
		queryErr error
	)

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 构建分页查询
	queryStmt := fn(SELECT(cols)).ORDER_BY(orderBy...).LIMIT(int64(size)).OFFSET(int64(offset))

	queryStart := time.Now()
	defer func() {
		if internal.Logger != nil {
			internal.Logger(ctx, internal.Minify(queryStmt.DebugSql()), time.Since(queryStart), queryErr)
		}
	}()

	queryErr = queryStmt.QueryContext(ctx, db, &dest)
	if queryErr != nil {
		return nil, 0, queryErr
	}
	return dest, total.Count, nil
}

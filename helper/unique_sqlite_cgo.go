//go:build cgo

package helper

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

func sqliteUniqueError(err error) (matched, unique bool) {
	isUnique := func(code sqlite3.ErrNoExtended) bool {
		return code == sqlite3.ErrConstraintUnique ||
			code == sqlite3.ErrConstraintPrimaryKey ||
			code == sqlite3.ErrConstraintRowID
	}

	if e, ok := errors.AsType[sqlite3.Error](err); ok {
		return true, isUnique(e.ExtendedCode)
	}

	if e, ok := errors.AsType[*sqlite3.Error](err); ok {
		return true, e != nil && isUnique(e.ExtendedCode)
	}
	return false, false
}

//go:build cgo

package helper

import (
	"fmt"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueDuplicateErrorSQLite(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"unique", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintUnique}, true},
		{"primary", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintPrimaryKey}, true},
		{"rowid", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintRowID}, true},
		{"pointer", &sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintUnique}, true},
		{"foreign_key", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintForeignKey}, false},
		{"not_null", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintNotNull}, false},
		{"check", sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintCheck}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsUniqueDuplicateError(tc.err))
			wrapped := fmt.Errorf("transaction: %w", fmt.Errorf("insert: %w", tc.err))
			assert.Equal(t, tc.want, IsUniqueDuplicateError(wrapped), "wrapped error")
		})
	}
}

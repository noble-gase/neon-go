package helper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{"url": "https://github.com/noble-gase?id=996&name=og"}

	b, err := MarshalNoEscapeHTML(data)
	assert.Nil(t, err)
	assert.Equal(t, string(b), `{"url":"https://github.com/noble-gase?id=996&name=og"}`)
}

func TestVersionCompare(t *testing.T) {
	ok, err := VersionCompare("1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.4")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">2.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "3.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestIsUniqueDuplicateError(t *testing.T) {
	assert.False(t, IsUniqueDuplicateError(nil))
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"mysql", &mysql.MySQLError{Number: 1062}, true},
		{"postgres", &pgconn.PgError{Code: "23505"}, true},
		{"mysql_foreign_key", &mysql.MySQLError{Number: 1452, Message: "Duplicate entry 'value' for key 'key_name'"}, false},
		{"postgres_foreign_key", &pgconn.PgError{Code: "23503", Message: "violates unique constraint"}, false},
		{"mysql_text", errors.New("ERROR 1062 (23000): Duplicate entry 'value' for key 'key_name'"), true},
		{"mysql_code_boundary", errors.New("Error 10620: another error"), false},
		{"mysql_code_suffix", errors.New("Error 1062abc"), false},
		{"mysql_text_without_code", errors.New("Duplicate entry 'value' for key 'key_name'"), true},
		{"postgres_text", errors.New(`duplicate key value violates unique constraint "constraint_name"`), true},
		{"sqlite_text", errors.New("UNIQUE constraint failed: table_name.column_name"), true},
		{"sqlite_primary_text", errors.New("PRIMARY KEY constraint failed: table_name.id"), true},
		{"other_error", errors.New("connection refused"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsUniqueDuplicateError(tc.err))
			wrapped := fmt.Errorf("transaction: %w", fmt.Errorf("insert: %w", tc.err))
			assert.Equal(t, tc.want, IsUniqueDuplicateError(wrapped), "wrapped error")
		})
	}
}

func TestExcelColumnIndex(t *testing.T) {
	assert.Equal(t, 0, ExcelColumnIndex("A"))
	assert.Equal(t, 1, ExcelColumnIndex("B"))
	assert.Equal(t, 26, ExcelColumnIndex("AA"))
	assert.Equal(t, 27, ExcelColumnIndex("AB"))
}

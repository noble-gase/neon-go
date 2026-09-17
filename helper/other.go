package helper

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"
)

// Nonce 生成随机串(size应为偶数)
func Nonce(size uint8) string {
	nonce := make([]byte, size/2)
	_, _ = io.ReadFull(rand.Reader, nonce)
	return hex.EncodeToString(nonce)
}

// MarshalNoEscapeHTML 不带HTML转义的JSON序列化
func MarshalNoEscapeHTML(v any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// 去掉 go std 给末尾加的 '\n'
	// @see https://github.com/golang/go/issues/7767
	b := buf.Bytes()
	if l := len(b); l != 0 && b[l-1] == '\n' {
		b = b[:l-1]
	}
	return b, nil
}

// VersionCompare 语义化的版本比较
//
//	支持：
//		>
//		>=
//		=
//		!=
//		<
//		<=
//		| (or)
//		& (and)
//
//	示例(rangeVer)：
//		1.0.0
//		=1.0.0
//		>2.0.0
//	    !=4.0.4
//		>=1.0.0&<2.0.0
//		<2.0.0|>3.0.0
func VersionCompare(rangeVer, curVer string) (bool, error) {
	semVer, err := version.NewVersion(curVer)
	if err != nil {
		return false, err
	}

	orVers := strings.SplitSeq(rangeVer, "|")
	for ver := range orVers {
		andVers := strings.Split(ver, "&")
		constraints, err := version.NewConstraint(strings.Join(andVers, ","))
		if err != nil {
			return false, err
		}
		if constraints.Check(semVer) {
			return true, nil
		}
	}
	return false, nil
}

// IsUniqueDuplicateError 判断是否为唯一索引或主键冲突，支持 %w 包装。
// 已知驱动按错误码判断，其他错误使用文本兜底。
func IsUniqueDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL 错误码判断
	if e, ok := errors.AsType[*mysql.MySQLError](err); ok {
		return e != nil && e.Number == 1062
	}

	// PostgreSQL 错误码判断
	var stateErr interface{ SQLState() string }
	if errors.As(err, &stateErr) {
		return stateErr.SQLState() == "23505"
	}

	// SQLite 错误码判断
	if matched, unique := sqliteUniqueError(err); matched {
		return unique
	}

	// 文本兜底判断
	msg := strings.ToLower(err.Error())
	for _, s := range []string{
		"duplicate entry",               // MySQL
		"violates unique constraint",    // PostgreSQL
		"unique constraint failed",      // SQLite
		"primary key constraint failed", // SQLite
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// ExcelColumnIndex 返回Excel列名对应的序号，如：A=0，B=1，AA=26，AB=27
func ExcelColumnIndex(name string) int {
	name = strings.ToUpper(name)
	if ok, err := regexp.MatchString(`^[A-Z]{1,2}$`, name); err != nil || !ok {
		return -1
	}

	index := 0
	for i, v := range name {
		if i != 0 {
			index = (index + 1) * 26
		}
		index += int(v - 'A')
	}
	return index
}

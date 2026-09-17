//go:build !cgo

package helper

func sqliteUniqueError(error) (matched, unique bool) {
	return false, false
}

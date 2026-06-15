package kvkit

import (
	"net/url"
	"sort"
	"strings"
)

// EmptyMode 值为空时的Encode模式
type EmptyMode int

const (
	Default EmptyMode = iota // 默认：bar=baz&foo=
	Ignore                   // 忽略：bar=baz
	OnlyKey                  // 仅保留Key：bar=baz&foo
)

// KV 用于处理 k/v 需要格式化的场景，如：签名
type KV map[string]string

// Set 设置 k/v
func (kv KV) Set(key, value string) {
	kv[key] = value
}

// Get 获取值
func (kv KV) Get(key string) string {
	return kv[key]
}

// Del 删除Key
func (kv KV) Del(key string) {
	delete(kv, key)
}

// Has 判断Key是否存在
func (kv KV) Has(key string) bool {
	_, ok := kv[key]
	return ok
}

// Encode 通过自定义的符号和分隔符按照key的ASCII码升序格式化为字符串。
// 例如：("=", "&") ---> bar=baz&foo=quux；
// 例如：(":", "#") ---> bar:baz#foo:quux；
func (kv KV) Encode(sym, sep string, emptyMode EmptyMode, ignoreKeys ...string) string {
	if len(kv) == 0 {
		return ""
	}

	ignoreKeyMap := make(map[string]struct{}, len(ignoreKeys))
	for _, k := range ignoreKeys {
		ignoreKeyMap[k] = struct{}{}
	}

	keys := make([]string, 0, len(kv))
	for k := range kv {
		if _, ok := ignoreKeyMap[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		val := kv[k]
		if len(val) == 0 && emptyMode == Ignore {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		buf.WriteString(k)

		if len(val) != 0 {
			buf.WriteString(sym)
			buf.WriteString(val)
			continue
		}

		// 保留符号
		if emptyMode != OnlyKey {
			buf.WriteString(sym)
		}
	}
	return buf.String()
}

// EncodeEscape 通过自定义的符号和分隔符按照key的ASCII码升序格式化为字符串。
// 例如：("=", "&") ---> bar=baz&foo=quux；
// 例如：(":", "#") ---> bar:baz#foo:quux；
func (kv KV) EncodeEscape(sym, sep string, emptyMode EmptyMode, ignoreKeys ...string) string {
	if len(kv) == 0 {
		return ""
	}

	ignoreKeyMap := make(map[string]struct{}, len(ignoreKeys))
	for _, k := range ignoreKeys {
		ignoreKeyMap[k] = struct{}{}
	}

	keys := make([]string, 0, len(kv))
	for k := range kv {
		if _, ok := ignoreKeyMap[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		val := kv[k]
		if len(val) == 0 && emptyMode == Ignore {
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString(sep)
		}

		buf.WriteString(url.QueryEscape(k))

		if len(val) != 0 {
			buf.WriteString(sym)
			buf.WriteString(url.QueryEscape(val))
			continue
		}

		// 保留符号
		if emptyMode != OnlyKey {
			buf.WriteString(sym)
		}
	}
	return buf.String()
}

// URLEncode encodes the KV into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
func (kv KV) URLEncode() string {
	query := url.Values{}
	for k, v := range kv {
		query.Set(k, v)
	}
	return query.Encode()
}

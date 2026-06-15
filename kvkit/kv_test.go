package kvkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	kv1 := KV{}
	kv1.Set("bar", "baz")
	kv1.Set("foo", "quux")

	assert.Equal(t, "bar=baz&foo=quux", kv1.Encode("=", "&", Default))
	assert.Equal(t, "bar:baz#foo:quux", kv1.Encode(":", "#", Default))

	kv2 := KV{}
	kv2.Set("hello", "world")
	kv2.Set("bar", "baz")
	kv2.Set("foo", "")

	assert.Equal(t, "bar=baz&foo=&hello=world", kv2.Encode("=", "&", Default))
	assert.Equal(t, "bar=baz&foo=&hello=world", kv2.Encode("=", "&", Default))
	assert.Equal(t, "bar=baz&foo&hello=world", kv2.Encode("=", "&", OnlyKey))
	assert.Equal(t, "bar=baz&hello=world", kv2.Encode("=", "&", Ignore))
	assert.Equal(t, "bar=baz&foo=", kv2.Encode("=", "&", Default, "hello"))
	assert.Equal(t, "bar=baz", kv2.Encode("=", "&", Ignore, "hello"))
}

func TestEncodeEscape(t *testing.T) {
	kv1 := KV{}
	kv1.Set("bar", "baz@666")
	kv1.Set("foo", "quux%666")

	assert.Equal(t, "bar=baz%40666&foo=quux%25666", kv1.EncodeEscape("=", "&", Default))
	assert.Equal(t, "bar:baz%40666#foo:quux%25666", kv1.EncodeEscape(":", "#", Default))

	kv2 := KV{}
	kv2.Set("hello", "world@666")
	kv2.Set("bar", "baz@666")
	kv2.Set("foo", "")

	assert.Equal(t, "bar=baz%40666&foo=&hello=world%40666", kv2.EncodeEscape("=", "&", Default))
	assert.Equal(t, "bar=baz%40666&foo=&hello=world%40666", kv2.EncodeEscape("=", "&", Default))
	assert.Equal(t, "bar=baz%40666&foo&hello=world%40666", kv2.EncodeEscape("=", "&", OnlyKey))
	assert.Equal(t, "bar=baz%40666&hello=world%40666", kv2.EncodeEscape("=", "&", Ignore))
	assert.Equal(t, "bar=baz%40666&foo=", kv2.EncodeEscape("=", "&", Default, "hello"))
	assert.Equal(t, "bar=baz%40666", kv2.EncodeEscape("=", "&", Ignore, "hello"))
}

func TestURLEncode(t *testing.T) {
	kv := KV{}
	kv.Set("bar", "baz@666")
	kv.Set("foo", "quux%666")

	assert.Equal(t, "bar=baz%40666&foo=quux%25666", kv.URLEncode())
}

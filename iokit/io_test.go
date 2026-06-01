package iokit

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitedWriter(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	w1 := LimitWriter(buf, 10)
	n, err := w1.Write([]byte("Hello"))
	assert.Equal(t, 5, n)
	assert.Nil(t, err)
	assert.Equal(t, "Hello", buf.String())

	buf.Reset()

	w2 := LimitWriter(buf, 5)
	n, err = w2.Write([]byte("Hello, world!"))
	assert.Equal(t, 5, n)
	assert.Nil(t, err)
	assert.Equal(t, "Hello", buf.String())
}

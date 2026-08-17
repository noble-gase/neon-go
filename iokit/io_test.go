package iokit

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNopWriteCloserReaderFromForwarding(t *testing.T) {
	for _, tc := range [...]struct {
		name string
		w    io.Writer
	}{
		{"not a ReaderFrom", io.Writer(nil)},
		{"a ReaderFrom", struct {
			io.Writer
			io.ReaderFrom
		}{}},
		{"bytes.Buffer", new(bytes.Buffer)},
	} {
		nwc := NopWriteCloser(tc.w)

		_, expected := tc.w.(io.ReaderFrom)
		_, got := nwc.(io.ReaderFrom)
		assert.Equal(t, expected, got, "NopWriteCloser incorrectly forwards ReaderFrom for %s", tc.name)
	}
}

func TestNopWriteCloserReadFrom(t *testing.T) {
	buf := new(bytes.Buffer)
	nwc := NopWriteCloser(buf)

	rf, ok := nwc.(io.ReaderFrom)
	require.True(t, ok)

	n, err := rf.ReadFrom(strings.NewReader("hello"))
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.Equal(t, "hello", buf.String())
}

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

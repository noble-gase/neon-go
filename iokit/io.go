package iokit

import "io"

// LimitWriter returns a Writer that writes to w
// but discards any bytes written beyond n bytes.
// The underlying implementation is a *LimitedWriter.
func LimitWriter(w io.Writer, n int64) io.Writer { return &LimitedWriter{w, n} }

// A LimitedWriter writes to W but limits the amount of
// data written to at most N bytes. Each call to Write
// updates N to reflect the remaining write quota.
// Write does not return an error when the limit is exceeded;
// instead, it discards any bytes beyond N.
type LimitedWriter struct {
	W io.Writer // underlying writer
	N int64     // max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, nil
	}
	if int64(len(p)) > l.N {
		p = p[:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}

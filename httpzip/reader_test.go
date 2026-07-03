package httpzip

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	r, err := NewReader(context.Background(), "https://ota.cdn.sunmi.com/OTA/2y5Fp5PPeEOnLFVdOyk9rw6va2p.zip")
	assert.Nil(t, err)

	buf := bytes.NewBuffer(nil)
	// 列出所有文件
	for _, f := range r.File {
		// 只取 version.txt
		if f.Name == "version.txt" {
			rc, _err := f.Open()
			assert.Nil(t, _err)
			defer rc.Close()

			io.Copy(buf, rc)
			break
		}
	}
	fmt.Println(buf.String())
}

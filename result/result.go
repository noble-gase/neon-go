package result

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/noble-gase/neon/codekit"
)

const MaxBufSize = 32 << 10 // 32KB

var bufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4<<10)) // 4KB
	},
}

// Result the result definition for API
type Result interface {
	// JSON outputs json result
	JSON(w http.ResponseWriter, r *http.Request)
}

type result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func (ret *result) JSON(w http.ResponseWriter, r *http.Request) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() > MaxBufSize {
			return
		}
		buf.Reset()
		bufPool.Put(buf)
	}()

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(ret); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func New(code codekit.Code, data ...any) Result {
	ret := &result{
		Code: code.Value(),
		Msg:  code.Message(),
	}
	if len(data) != 0 {
		ret.Data = data[0]
	}
	return ret
}

func OK(data ...any) Result {
	return New(codekit.OK, data...)
}

func Err(err error) Result {
	var code codekit.Code
	if errors.As(err, &code) {
		return New(code)
	}
	return New(codekit.Err)
}

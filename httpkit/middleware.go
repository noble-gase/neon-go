package httpkit

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/noble-gase/neon/helper"
	"github.com/noble-gase/neon/iokit"
	"github.com/noble-gase/neon/result"
	"github.com/tidwall/pretty"
	"google.golang.org/grpc/metadata"
)

const MaxLogSize = 32 << 10 // 32KB

var bufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4<<10)) // 4KB
	},
}

// Recovery is a middleware that recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(r.Context(), "panic occurred", slog.Any("error", err), slog.String("stack", string(debug.Stack())))
				result.Err(fmt.Errorf("panic occurred: %+v", err)).JSON(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Header is a middleware that injects http headers into the context of each request.
func Header(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.Pairs()
		}

		for k, vals := range r.Header {
			md.Set(k, vals...)
		}

		// traceId
		var traceId string
		if vals := md.Get(helper.XTraceID); len(vals) != 0 {
			traceId = vals[0]
		}
		if len(traceId) == 0 {
			traceId = strings.ReplaceAll(uuid.New().String(), "-", "")
			md.Set(helper.XTraceID, traceId)
		}

		// set response header
		w.Header().Set(helper.XTraceID, traceId)

		// reset request context
		next.ServeHTTP(w, r.WithContext(metadata.NewIncomingContext(ctx, md)))
	})
}

// Log is a middleware that logs the request and response
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		body := "<nil>"

		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer func() {
			if buf.Cap() > MaxLogSize {
				return
			}
			buf.Reset()
			bufPool.Put(buf)
		}()

		// 自定义响应
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		ww.Tee(iokit.LimitWriter(buf, MaxLogSize))

		defer func() {
			slog.InfoContext(
				r.Context(), r.RequestURI,
				slog.String("method", r.Method),
				slog.Any("header", r.Header),
				slog.String("request", body),
				slog.String("response", strings.TrimSuffix(buf.String(), "\n")),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", ww.Status()),
				slog.String("duration", time.Since(start).String()),
			)
		}()

		if r.Body != nil && r.Body != http.NoBody {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				slog.ErrorContext(r.Context(), "request body read failed", slog.String("error", err.Error()))
				result.Err(err).JSON(w, r)
				return
			}
			_ = r.Body.Close()

			if len(b) > MaxLogSize {
				body = "the body is discarded because it exceeds the maximum log size"
			} else {
				if ContentType(r.Header) == ContentJSON {
					body = string(pretty.Ugly(b))
				} else {
					body = string(b)
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(b))
		}

		next.ServeHTTP(ww, r)
	})
}

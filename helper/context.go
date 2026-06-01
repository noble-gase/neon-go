package helper

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
	"google.golang.org/grpc/metadata"
)

const XTraceId = "x-trace-id"

// CtxWithMDValue sets key-value pairs to the incoming metadata
// and returns a new context.
func CtxWithMDValue(ctx context.Context, key string, vals ...string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.Pairs()
	}
	md.Set(key, vals...)
	return metadata.NewIncomingContext(ctx, md)
}

// CtxWithTraceId ensures a trace ID exists in the incoming metadata.
// If absent, a new trace ID is generated and attached.
func CtxWithTraceId(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.Pairs()
	}
	if len(md.Get(XTraceId)) != 0 {
		return ctx
	}

	traceId := strings.ReplaceAll(uuid.New().String(), "-", "")

	md.Set(XTraceId, traceId)
	return metadata.NewIncomingContext(ctx, md)
}

// MDValFromCtx returns value for the given key from incoming metadata.
func MDValFromCtx(ctx context.Context, key string) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return md.Get(key)
}

// MDStrFromCtx returns the string value for the given key from incoming metadata.
func MDStrFromCtx(ctx context.Context, key string) string {
	ss := MDValFromCtx(ctx, key)
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}

// MDBoolFromCtx returns the boolean value for the given key from incoming metadata.
func MDBoolFromCtx(ctx context.Context, key string) bool {
	s := MDStrFromCtx(ctx, key)
	v, _ := strconv.ParseBool(s)
	return v
}

// MDIntFromCtx returns the signed integer value for the given key from incoming metadata.
func MDIntFromCtx[T constraints.Signed](ctx context.Context, key string) T {
	s := MDStrFromCtx(ctx, key)
	v, _ := strconv.ParseInt(s, 10, 64)
	return T(v)
}

// MDUintFromCtx returns the unsigned integer value for the given key from incoming metadata.
func MDUintFromCtx[T constraints.Unsigned](ctx context.Context, key string) T {
	s := MDStrFromCtx(ctx, key)
	v, _ := strconv.ParseUint(s, 10, 64)
	return T(v)
}

// MDFloatFromCtx returns the float value for the given key from incoming metadata.
func MDFloatFromCtx[T constraints.Float](ctx context.Context, key string) T {
	s := MDStrFromCtx(ctx, key)
	v, _ := strconv.ParseFloat(s, 64)
	return T(v)
}

// MDTraceIdFromCtx returns the trace ID from incoming metadata.
func MDTraceIdFromCtx(ctx context.Context) string {
	return MDStrFromCtx(ctx, XTraceId)
}

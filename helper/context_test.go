package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestCtxWithMDValue(t *testing.T) {
	ctx := CtxWithMDValue(context.Background(), "user", "noble-gase")
	md, ok := metadata.FromIncomingContext(ctx)
	assert.True(t, ok)
	value := md.Get("user")
	assert.Equal(t, []string{"noble-gase"}, value)
}

func TestCtxWithTraceId(t *testing.T) {
	ctx1 := CtxWithTraceID(context.Background())
	md, ok := metadata.FromIncomingContext(ctx1)
	assert.True(t, ok)
	value := md.Get(XTraceID)
	assert.Equal(t, 1, len(value))

	ctx := CtxWithMDValue(context.Background(), XTraceID, "noble-gase")
	ctx2 := CtxWithTraceID(ctx)
	md, ok = metadata.FromIncomingContext(ctx2)
	assert.True(t, ok)
	value = md.Get(XTraceID)
	assert.Equal(t, []string{"noble-gase"}, value)
}

func TestMDValueFromCtx(t *testing.T) {
	ctx := CtxWithMDValue(context.Background(), "user", "noble-gase")
	assert.Equal(t, []string{"noble-gase"}, MDValFromCtx(ctx, "user"))
}

func TestMDStrFromCtx(t *testing.T) {
	ctx := CtxWithMDValue(context.Background(), "user", "noble-gase")
	assert.Equal(t, "noble-gase", MDStrFromCtx(ctx, "user"))
}

func TestMDBoolFromCtx(t *testing.T) {
	assert.False(t, MDBoolFromCtx(context.Background(), "status"))

	ctx := CtxWithMDValue(context.Background(), "status", "true")
	assert.True(t, MDBoolFromCtx(ctx, "status"))
}

func TestMDIntFromCtx(t *testing.T) {
	assert.Equal(t, 0, MDIntFromCtx[int](context.Background(), "total"))

	ctx := CtxWithMDValue(context.Background(), "total", "1")
	assert.Equal(t, 1, MDIntFromCtx[int](ctx, "total"))
}

func TestMDUintFromCtx(t *testing.T) {
	assert.Equal(t, uint(0), MDUintFromCtx[uint](context.Background(), "total"))

	ctx := CtxWithMDValue(context.Background(), "total", "1")
	assert.Equal(t, uint(1), MDUintFromCtx[uint](ctx, "total"))
}

func TestMDFloatFromCtx(t *testing.T) {
	assert.Equal(t, float64(0), MDFloatFromCtx[float64](context.Background(), "rate"))

	ctx := CtxWithMDValue(context.Background(), "rate", "3.14")
	assert.Equal(t, float64(3.14), MDFloatFromCtx[float64](ctx, "rate"))
}

func TestMDTraceIdFromCtx(t *testing.T) {
	ctx1 := CtxWithTraceID(context.Background())
	assert.NotEqual(t, 0, len(MDTraceIDFromCtx(ctx1)))

	ctx := CtxWithMDValue(context.Background(), XTraceID, "noble-gase")
	ctx2 := CtxWithTraceID(ctx)
	assert.Equal(t, "noble-gase", MDTraceIDFromCtx(ctx2))
}

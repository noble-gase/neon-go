package redkit

import (
	"context"
	"errors"
	"testing"
)

func TestDoSFContextCanceled(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		_, _ = doSF[string](context.Background(), "context-canceled", func() (any, error) {
			close(started)
			<-release
			return "value", nil
		})
	}()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	value, err := doSF[string](ctx, "context-canceled", func() (any, error) {
		t.Fatal("shared function should not run twice")
		return "", nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if value != "" {
		t.Fatalf("expected zero value, got %q", value)
	}

	close(release)
	<-done
}

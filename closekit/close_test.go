package closekit

import (
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
)

func TestClose(t *testing.T) {
	resetClosers(t)

	var closed []string
	add := func(name string, priority Priority, err error) {
		Add(name, priority, func() error {
			closed = append(closed, name)
			return err
		})
	}

	add("db", P10, nil)
	add("worker-1", P8, errors.New("close failed"))
	add("mq", P7, nil)
	add("worker-2", P8, nil)

	Close()

	want := []string{"mq", "worker-1", "worker-2", "db"}
	if !reflect.DeepEqual(closed, want) {
		t.Fatalf("close order = %v, want %v", closed, want)
	}

	Close()
	if !reflect.DeepEqual(closed, want) {
		t.Fatalf("second Close() closed resources again: got %v", closed)
	}
}

func TestClose_AddFromCloseFunc(t *testing.T) {
	resetClosers(t)

	var closed []string
	Add("first", P0, func() error {
		closed = append(closed, "first")
		Add("next", P0, func() error {
			closed = append(closed, "next")
			return nil
		})
		return nil
	})

	Close()
	if want := []string{"first"}; !reflect.DeepEqual(closed, want) {
		t.Fatalf("first Close() = %v, want %v", closed, want)
	}

	Close()
	if want := []string{"first", "next"}; !reflect.DeepEqual(closed, want) {
		t.Fatalf("second Close() = %v, want %v", closed, want)
	}
}

func resetClosers(t *testing.T) {
	t.Helper()

	mutex.Lock()
	closers = nil
	mutex.Unlock()

	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() {
		mutex.Lock()
		closers = nil
		mutex.Unlock()
		slog.SetDefault(originalLogger)
	})
}

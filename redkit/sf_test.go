package redkit

import (
	"errors"
	"strings"
	"testing"
)

func TestDoSF(t *testing.T) {
	wantErr := errors.New("load failed")

	tests := []struct {
		name    string
		key     string
		fn      func() (any, error)
		want    string
		wantErr error
	}{
		{
			name: "success",
			key:  "success",
			fn: func() (any, error) {
				return "value", nil
			},
			want: "value",
		},
		{
			name: "error",
			key:  "error",
			fn: func() (any, error) {
				return nil, wantErr
			},
			wantErr: wantErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := doSF[string](test.key, test.fn)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if value != test.want {
				t.Fatalf("value = %q, want %q", value, test.want)
			}
		})
	}

	_, err := doSF[string]("unexpected-type", func() (any, error) {
		return 1, nil
	})
	if err == nil || !strings.Contains(err.Error(), "unexpected result type int") {
		t.Fatalf("unexpected type error = %v", err)
	}
}

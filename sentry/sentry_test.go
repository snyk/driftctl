package sentry

import (
	"testing"

	cmderrors "github.com/cloudskiff/driftctl/pkg/cmd/errors"
	"github.com/pkg/errors"
)

func Test_shouldCaptureException(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "should not capture errors.UsageError",
			err:  cmderrors.UsageError{},
			want: false,
		},
		{
			name: "should not capture wrapped errors.UsageError",
			err:  errors.Wrap(cmderrors.UsageError{}, "test"),
			want: false,
		},
		{
			name: "should capture errors.withStack",
			err:  errors.New("test"),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCaptureException(tt.err); got != tt.want {
				t.Errorf("shouldCaptureException() = %v, want %v", got, tt.want)
			}
		})
	}
}

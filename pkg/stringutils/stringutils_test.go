package stringutils

import (
	"testing"
)

func TestUnescape(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "\\ unescaped at start",
			line: "\\",
			want: "",
		},
		{
			name: "\\ escaped at start",
			line: "\\\\",
			want: "\\",
		},
		{
			name: "dot escaped",
			line: "\\.this\\.is\\.dotted\\.",
			want: ".this.is.dotted.",
		},
		{
			name: "mix escaped",
			line: "\\.this\\.is\\\\.dotted\\.",
			want: ".this.is\\.dotted.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unescape(tt.line); got != tt.want {
				t.Errorf("Unescape() = %v, want %v", got, tt.want)
			}
		})
	}
}

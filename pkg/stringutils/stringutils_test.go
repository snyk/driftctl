package stringutils

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "Dot at start",
			line: ".",
			want: []string{"."},
		},
		{
			name: "Dot at end",
			line: "test.toto.",
			want: []string{"test", "toto"},
		},
		{
			name: "wildcard dot",
			line: "*.subfoobar",
			want: []string{"*", "subfoobar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapableSplit(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscapableSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

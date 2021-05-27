package aws

import (
	"reflect"
	"testing"
)

func TestAwsInstance_Attrs(t *testing.T) {
	tests := []struct {
		name     string
		instance *AwsInstance
		want     map[string]string
	}{
		{
			name:     "empty attrs",
			instance: &AwsInstance{},
			want:     map[string]string{},
		},
		{
			name: "Tag name",
			instance: &AwsInstance{
				Tags: map[string]string{"Name": "name"},
			},
			want: map[string]string{
				"Name": "name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.instance
			if got := r.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

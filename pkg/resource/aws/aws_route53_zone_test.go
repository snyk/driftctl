package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53Zone_String(t *testing.T) {
	tests := []struct {
		name string
		zone AwsRoute53Zone
		want string
	}{
		{name: "",
			zone: AwsRoute53Zone{
				Name: aws.String("example.com"),
			},
			want: "example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.zone.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

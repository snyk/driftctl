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
		{name: "test route53 zone stringer with name and id",
			zone: AwsRoute53Zone{
				Name: aws.String("example.com"),
				Id:   "Z04218102KCRRR1DWDYJT",
			},
			want: "example.com (Id: Z04218102KCRRR1DWDYJT)",
		},
		{name: "test route53 zone stringer without name",
			zone: AwsRoute53Zone{
				Name: nil,
				Id:   "Z04218102KCRRR1DWDYJT",
			},
			want: "Z04218102KCRRR1DWDYJT",
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

package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53Zone_Attrs(t *testing.T) {
	tests := []struct {
		name string
		zone AwsRoute53Zone
		want map[string]string
	}{
		{name: "test route53 zone attrs with name",
			zone: AwsRoute53Zone{
				Name: aws.String("example.com"),
			},
			want: map[string]string{
				"Name": "example.com",
			},
		},
		{name: "test route53 zone attrs without name",
			zone: AwsRoute53Zone{
				Name: nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.zone.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

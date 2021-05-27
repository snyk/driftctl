package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsIamAccessKey_Attrs(t *testing.T) {
	tests := []struct {
		user   string
		access AwsIamAccessKey
		want   map[string]string
	}{
		{user: "test iam access key attrs with user",
			access: AwsIamAccessKey{
				User: aws.String("test_user"),
			},
			want: map[string]string{
				"User": "test_user",
			},
		},
		{user: "test iam access key attrs without user",
			access: AwsIamAccessKey{
				User: nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.user, func(t *testing.T) {
			if got := tt.access.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

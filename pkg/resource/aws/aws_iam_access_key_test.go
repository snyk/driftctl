package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsIamAccessKey_String(t *testing.T) {
	tests := []struct {
		user   string
		access AwsIamAccessKey
		want   string
	}{
		{user: "test iam access key stringer with user and id",
			access: AwsIamAccessKey{
				User: aws.String("test_user"),
				Id:   "AKIA2SIQ53JH4CMB42VB",
			},
			want: "AKIA2SIQ53JH4CMB42VB (User: test_user)",
		},
		{user: "test iam access key stringer without user",
			access: AwsIamAccessKey{
				User: nil,
				Id:   "AKIA2SIQ53JH4CMB42VB",
			},
			want: "AKIA2SIQ53JH4CMB42VB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.user, func(t *testing.T) {
			if got := tt.access.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

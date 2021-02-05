package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsSnsTopic_String(t *testing.T) {
	tests := []struct {
		name     string
		snsTopic AwsSnsTopic
		want     string
	}{
		{
			name: "DisplayName and Name not nil",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String("[DisplayName]"),
				Id:          "[ID]",
				Name:        aws.String("[Name]"),
			},
			want: "[DisplayName] ([Name])",
		},
		{
			name: "DisplayName empty and Name not empty",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String(""),
				Id:          "[ID]",
				Name:        aws.String("[Name]"),
			},
			want: "[Name]",
		},
		{
			name: "DisplayName and Name empty",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String(""),
				Id:          "[ID]",
				Name:        aws.String(""),
			},
			want: "[ID]",
		},
		{
			name: "DisplayName and Name nil",
			snsTopic: AwsSnsTopic{
				DisplayName: nil,
				Id:          "[ID]",
				Name:        nil,
			},
			want: "[ID]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.snsTopic.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

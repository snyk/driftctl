package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsSnsTopic_Attrs(t *testing.T) {
	tests := []struct {
		name     string
		snsTopic AwsSnsTopic
		want     map[string]string
	}{
		{
			name: "DisplayName and Name not nil",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String("[DisplayName]"),
				Name:        aws.String("[Name]"),
			},
			want: map[string]string{
				"DisplayName": "[DisplayName]",
				"Name":        "[Name]",
			},
		},
		{
			name: "DisplayName not empty and Name empty",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String(""),
			},
			want: map[string]string{},
		},
		{
			name: "DisplayName empty and Name not empty",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String(""),
				Name:        aws.String("[Name]"),
			},
			want: map[string]string{
				"Name": "[Name]",
			},
		},
		{
			name: "DisplayName and Name empty",
			snsTopic: AwsSnsTopic{
				DisplayName: aws.String(""),
				Name:        aws.String(""),
			},
			want: map[string]string{},
		},
		{
			name: "DisplayName and Name nil",
			snsTopic: AwsSnsTopic{
				DisplayName: nil,
				Name:        nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.snsTopic.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

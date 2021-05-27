package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsLambdaEventSourceMapping_Attrs(t *testing.T) {
	tests := []struct {
		name   string
		lambda AwsLambdaEventSourceMapping
		want   map[string]string
	}{
		{name: "test lambda attrs with source and dest",
			lambda: AwsLambdaEventSourceMapping{
				EventSourceArn: aws.String("source-arn"),
				FunctionName:   aws.String("function-name"),
			},
			want: map[string]string{
				"Source": "source-arn",
				"Dest":   "function-name",
			},
		},
		{name: "test lambda attrs with source",
			lambda: AwsLambdaEventSourceMapping{
				EventSourceArn: aws.String("source-arn"),
			},
			want: map[string]string{},
		},
		{name: "test lambda attrs without values",
			lambda: AwsLambdaEventSourceMapping{},
			want:   map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lambda.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

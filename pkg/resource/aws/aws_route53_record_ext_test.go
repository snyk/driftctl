package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53Record_Attrs(t *testing.T) {
	tests := []struct {
		name   string
		record AwsRoute53Record
		want   map[string]string
	}{
		{name: "test route53 record attrs with name and fqdn and type and zoneId",
			record: AwsRoute53Record{
				Name:   aws.String("example.com"),
				Fqdn:   aws.String("_github-challenge-cloudskiff.cloudskiff.com"),
				Type:   aws.String("TXT"),
				ZoneId: aws.String("ZOS30SFDAFTU9"),
			},
			want: map[string]string{
				"Fqdn":   "_github-challenge-cloudskiff.cloudskiff.com",
				"Type":   "TXT",
				"ZoneId": "ZOS30SFDAFTU9",
			},
		},
		{name: "test route53 record attrs without values",
			record: AwsRoute53Record{
				Name: nil,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

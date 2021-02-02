package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53Record_String(t *testing.T) {
	tests := []struct {
		name   string
		record AwsRoute53Record
		want   string
	}{
		{name: "test route53 record stringer with name and fqdn and type and zoneId",
			record: AwsRoute53Record{
				Name:   aws.String("example.com"),
				Fqdn:   aws.String("_github-challenge-cloudskiff.cloudskiff.com"),
				Type:   aws.String("TXT"),
				ZoneId: aws.String("ZOS30SFDAFTU9"),
				Id:     "ZOS30SFDAFTU9__github-challenge-cloudskiff.cloudskiff.com_TXT",
			},
			want: "_github-challenge-cloudskiff.cloudskiff.com (TXT) (Zone: ZOS30SFDAFTU9)",
		},
		{name: "test route53 record stringer without values",
			record: AwsRoute53Record{
				Name: nil,
				Id:   "ZOS30SFDAFTU9__github-challenge-cloudskiff.cloudskiff.com_TXT",
			},
			want: "ZOS30SFDAFTU9__github-challenge-cloudskiff.cloudskiff.com_TXT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

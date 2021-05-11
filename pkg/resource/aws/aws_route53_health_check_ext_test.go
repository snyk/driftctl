package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53HealthCheck_Attrs(t *testing.T) {
	tests := []struct {
		name        string
		healthCheck *AwsRoute53HealthCheck
		want        map[string]string
	}{
		{
			name:        "empty attrs",
			healthCheck: &AwsRoute53HealthCheck{},
			want:        map[string]string{},
		},
		{
			name: "Tag name with fqdn and respath",
			healthCheck: &AwsRoute53HealthCheck{
				Tags:         map[string]string{"Name": "name"},
				Fqdn:         aws.String("fq.dn"),
				ResourcePath: aws.String("/toto"),
			},
			want: map[string]string{
				"Name": "name",
				"Fqdn": "fq.dn",
				"Path": "/toto",
			},
		},
		{
			name: "Tag name with ip and port",
			healthCheck: &AwsRoute53HealthCheck{
				Tags:      map[string]string{"Name": "name"},
				IpAddress: aws.String("10.0.0.10"),
				Port:      aws.Int(443),
			},
			want: map[string]string{
				"Name":      "name",
				"IpAddress": "10.0.0.10",
				"Port":      "443",
			},
		},
		{
			name: "Tag name with ip, port ans respath",
			healthCheck: &AwsRoute53HealthCheck{
				Tags:         map[string]string{"Name": "name"},
				IpAddress:    aws.String("10.0.0.10"),
				Port:         aws.Int(443),
				ResourcePath: aws.String("/toto"),
			},
			want: map[string]string{
				"Name":      "name",
				"IpAddress": "10.0.0.10",
				"Port":      "443",
				"Path":      "/toto",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.healthCheck
			if got := r.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

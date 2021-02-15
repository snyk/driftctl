package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute53HealthCheck_String(t *testing.T) {
	tests := []struct {
		name        string
		healthCheck *AwsRoute53HealthCheck
		want        string
	}{
		{
			name:        "Just id",
			healthCheck: &AwsRoute53HealthCheck{Id: "id"},
			want:        "id",
		},
		{
			name: "Tag name with fqdn and respath",
			healthCheck: &AwsRoute53HealthCheck{
				Id:           "id",
				Tags:         map[string]string{"Name": "name"},
				Fqdn:         aws.String("fq.dn"),
				ResourcePath: aws.String("/toto"),
			},
			want: "name (fqdn: fq.dn, path: /toto)",
		},
		{
			name: "Tag name with ip and port",
			healthCheck: &AwsRoute53HealthCheck{
				Id:        "id",
				Tags:      map[string]string{"Name": "name"},
				IpAddress: aws.String("10.0.0.10"),
				Port:      aws.Int(443),
			},
			want: "name (ip: 10.0.0.10, port: 443)",
		},
		{
			name: "Tag name with ip, port ans respath",
			healthCheck: &AwsRoute53HealthCheck{
				Id:           "id",
				Tags:         map[string]string{"Name": "name"},
				IpAddress:    aws.String("10.0.0.10"),
				Port:         aws.Int(443),
				ResourcePath: aws.String("/toto"),
			},
			want: "name (ip: 10.0.0.10, port: 443, path: /toto)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.healthCheck
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

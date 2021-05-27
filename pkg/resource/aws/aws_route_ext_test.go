package aws

import (
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestAwsRoute_Attrs(t *testing.T) {
	type fields struct {
		Id                       string
		RouteTableId             *string
		DestinationCidrBlock     *string
		DestinationIpv6CidrBlock *string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "test for cidr block",
			fields: fields{
				RouteTableId:         awssdk.String("table-id"),
				DestinationCidrBlock: awssdk.String("0.0.0.0/0"),
			},
			want: map[string]string{
				"Table":       "table-id",
				"Destination": "0.0.0.0/0",
			},
		},
		{
			name: "test for ipv6 cidr block",
			fields: fields{
				RouteTableId:             awssdk.String("table-id"),
				DestinationIpv6CidrBlock: awssdk.String("::/0"),
			},
			want: map[string]string{
				"Table":       "table-id",
				"Destination": "::/0",
			},
		},
		{
			name: "test without destination",
			fields: fields{
				RouteTableId: awssdk.String("table-id"),
			},
			want: map[string]string{
				"Table": "table-id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AwsRoute{
				Id:                       tt.fields.Id,
				RouteTableId:             tt.fields.RouteTableId,
				DestinationCidrBlock:     tt.fields.DestinationCidrBlock,
				DestinationIpv6CidrBlock: tt.fields.DestinationIpv6CidrBlock,
			}
			if got := r.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

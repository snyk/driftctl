package aws

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

func TestAwsRouteTableAssociation_String(t *testing.T) {
	type fields struct {
		GatewayId    *string
		Id           string
		RouteTableId *string
		SubnetId     *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test for gateway",
			fields: fields{
				GatewayId:    awssdk.String("gateway-id"),
				RouteTableId: awssdk.String("table-id"),
			},
			want: "Table: table-id, Gateway: gateway-id",
		},
		{
			name: "test for subnet",
			fields: fields{
				SubnetId:     awssdk.String("subnet-id"),
				RouteTableId: awssdk.String("table-id"),
			},
			want: "Table: table-id, Subnet: subnet-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AwsRouteTableAssociation{
				GatewayId:    tt.fields.GatewayId,
				Id:           tt.fields.Id,
				RouteTableId: tt.fields.RouteTableId,
				SubnetId:     tt.fields.SubnetId,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

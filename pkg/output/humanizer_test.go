package output

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	testresource "github.com/cloudskiff/driftctl/test/resource"
)

func TestHumanizeAttribute_AWS(t *testing.T) {
	tests := []struct {
		name string
		res  AttributesGetter
		want string
	}{
		{
			name: "test empty iam_access_key",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsIamAccessKeyResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test valid iam_access_key",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsIamAccessKeyResourceType,
				Attrs: &resource.Attributes{
					"user": "foo",
				},
			},
			want: "User: foo",
		},
		{
			name: "test empty aws_instance",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsInstanceResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test valid aws_instance",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsInstanceResourceType,
				Attrs: &resource.Attributes{
					"tags": map[string]interface{}{
						"name": "foo",
					},
				},
			},
			want: "Name: foo",
		},
		{
			name: "test empty aws_lambda_event_source_mapping",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsLambdaEventSourceMappingResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with source aws_lambda_event_source_mapping",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsLambdaEventSourceMappingResourceType,
				Attrs: &resource.Attributes{
					"event_source_arn": "source-arn",
				},
			},
			want: "",
		},
		{
			name: "test with source and dest aws_lambda_event_source_mapping",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsLambdaEventSourceMappingResourceType,
				Attrs: &resource.Attributes{
					"event_source_arn": "source-arn",
					"function_name":    "function-name",
				},
			},
			want: "Dest: function-name, Source: source-arn",
		},
		{
			name: "test empty aws_route",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsRouteResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with no destination aws_route",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRouteResourceType,
				Attrs: &resource.Attributes{
					"route_table_id": "table-id",
				},
			},
			want: "Table: table-id",
		},
		{
			name: "test with ipv4 destination aws_route",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRouteResourceType,
				Attrs: &resource.Attributes{
					"destination_cidr_block": "0.0.0.0/0",
					"route_table_id":         "table-id",
				},
			},
			want: "Destination: 0.0.0.0/0, Table: table-id",
		},
		{
			name: "test with ipv6 destination aws_route",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRouteResourceType,
				Attrs: &resource.Attributes{
					"destination_ipv6_cidr_block": "::/0",
					"route_table_id":              "table-id",
				},
			},
			want: "Destination: ::/0, Table: table-id",
		},
		{
			name: "test empty aws_route53_health_check",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsRoute53HealthCheckResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with name tag, fqdn and resource path aws_route53_health_check",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRoute53HealthCheckResourceType,
				Attrs: &resource.Attributes{
					"tags": map[string]interface{}{
						"name": "foo",
					},
					"fqdn":          "fq.dn",
					"resource_path": "/toto",
				},
			},
			want: "Fqdn: fq.dn, Name: foo, Path: /toto",
		},
		{
			name: "test with ip and port aws_route53_health_check",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRoute53HealthCheckResourceType,
				Attrs: &resource.Attributes{
					"tags": map[string]interface{}{
						"name": "foo",
					},
					"ip_address": "10.0.0.10",
					"port":       float64(443),
				},
			},
			want: "IpAddress: 10.0.0.10, Name: foo, Port: 443",
		},
		{
			name: "test with ip, port and resource path aws_route53_health_check",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRoute53HealthCheckResourceType,
				Attrs: &resource.Attributes{
					"tags": map[string]interface{}{
						"name": "foo",
					},
					"ip_address":    "10.0.0.10",
					"port":          float64(443),
					"resource_path": "/toto",
				},
			},
			want: "IpAddress: 10.0.0.10, Name: foo, Path: /toto, Port: 443",
		},
		{
			name: "test empty aws_route53_record",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsRoute53RecordResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with fqdn, type and zoneId aws_route53_record",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRoute53RecordResourceType,
				Attrs: &resource.Attributes{
					"fqdn":    "_github-challenge-cloudskiff.cloudskiff.com",
					"type":    "TXT",
					"zone_id": "ZOS30SFDAFTU9",
				},
			},
			want: "Fqdn: _github-challenge-cloudskiff.cloudskiff.com, Type: TXT, ZoneId: ZOS30SFDAFTU9",
		},
		{
			name: "test empty aws_route53_zone",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsRoute53ZoneResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with name aws_route53_zone",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRoute53ZoneResourceType,
				Attrs: &resource.Attributes{
					"name": "example.com",
				},
			},
			want: "Name: example.com",
		},
		{
			name: "test empty aws_route_table_association",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsRouteTableAssociationResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with gateway aws_route_table_association",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRouteTableAssociationResourceType,
				Attrs: &resource.Attributes{
					"route_table_id": "table-id",
					"gateway_id":     "gtw-id",
				},
			},
			want: "Gateway: gtw-id, Table: table-id",
		},
		{
			name: "test with subnet aws_route_table_association",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsRouteTableAssociationResourceType,
				Attrs: &resource.Attributes{
					"route_table_id": "table-id",
					"subnet_id":      "subnet-id",
				},
			},
			want: "Subnet: subnet-id, Table: table-id",
		},
		{
			name: "test empty aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test ingress_ssh_ipv4 aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "ingress",
					"protocol":          "tcp",
					"from_port":         float64(22),
					"to_port":           float64(22),
					"cidr_blocks":       []interface{}{"0.0.0.0/0", "1.2.3.4/32"},
				},
			},
			want: "Ports: 22, Protocol: tcp, SecurityGroup: sg-12345, Source: 0.0.0.0/0, 1.2.3.4/32, Type: ingress",
		},
		{
			name: "test egress_ssh_ipv4 aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "egress",
					"protocol":          "tcp",
					"from_port":         float64(22),
					"to_port":           float64(22),
					"cidr_blocks":       []interface{}{"0.0.0.0/0", "1.2.3.4/32"},
				},
			},
			want: "Destination: 0.0.0.0/0, 1.2.3.4/32, Ports: 22, Protocol: tcp, SecurityGroup: sg-12345, Type: egress",
		},
		{
			name: "test ingress_all aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "ingress",
					"protocol":          "-1",
				},
			},
			want: "Protocol: All, SecurityGroup: sg-12345, Type: ingress",
		},
		{
			name: "test ingress_all_range_0 aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "ingress",
					"protocol":          "-1",
					"from_port":         float64(0),
					"to_port":           float64(0),
				},
			},
			want: "Ports: All, Protocol: All, SecurityGroup: sg-12345, Type: ingress",
		},
		{
			name: "test ingress_all_ipv6 aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "ingress",
					"protocol":          "-1",
					"ipv6_cidr_blocks":  []interface{}{"::/0"},
				},
			},
			want: "Protocol: All, SecurityGroup: sg-12345, Source: ::/0, Type: ingress",
		},
		{
			name: "test ingress_all_prefix aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id": "sg-12345",
					"type":              "ingress",
					"protocol":          "-1",
					"prefix_list_ids":   []interface{}{"pl-12345"},
				},
			},
			want: "Protocol: All, SecurityGroup: sg-12345, Source: pl-12345, Type: ingress",
		},
		{
			name: "test egress_all_source aws_security_group_rule",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSecurityGroupRuleResourceType,
				Attrs: &resource.Attributes{
					"security_group_id":        "sg-12345",
					"type":                     "egress",
					"protocol":                 "all",
					"source_security_group_id": "sg-67890",
				},
			},
			want: "Destination: sg-67890, Protocol: all, SecurityGroup: sg-12345, Type: egress",
		},
		{
			name: "test empty aws_sns_topic",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  aws.AwsSnsTopicResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "",
		},
		{
			name: "test with name and display name aws_sns_topic",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: aws.AwsSnsTopicResourceType,
				Attrs: &resource.Attributes{
					"name":         "foo",
					"display_name": "bar",
				},
			},
			want: "DisplayName: bar, Name: foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)

			abstractResource, ok := tt.res.(*resource.AbstractResource)
			if ok {
				schema, _ := repo.GetSchema(abstractResource.TerraformType())
				abstractResource.Sch = schema
			}

			if got := HumanizeAttribute(tt.res); got != tt.want {
				t.Errorf("HumanizeAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHumanizeAttribute_Github(t *testing.T) {
	tests := []struct {
		name string
		res  AttributesGetter
		want string
	}{
		{
			name: "test empty github_branch_protection",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  github.GithubBranchProtectionResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "Id: foo",
		},
		{
			name: "test with pattern github_branch_protection",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: github.GithubBranchProtectionResourceType,
				Attrs: &resource.Attributes{
					"pattern": "my-branch",
				},
			},
			want: "Branch: my-branch, Id: foo",
		},
		{
			name: "test with pattern and invalid base64 repo_id github_branch_protection",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: github.GithubBranchProtectionResourceType,
				Attrs: &resource.Attributes{
					"pattern":       "my-branch",
					"repository_id": "invalid",
				},
			},
			want: "Branch: my-branch, Id: foo",
		},
		{
			name: "test with pattern and valid base64 repo_id github_branch_protection",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: github.GithubBranchProtectionResourceType,
				Attrs: &resource.Attributes{
					"pattern":       "my-branch",
					"repository_id": "MDEwOlJlcG9zaXRvcnkxMjM0NTY=",
				},
			},
			want: "Branch: my-branch, RepoId: 010:Repository123456",
		},
		{
			name: "test empty github_team",
			res: &resource.AbstractResource{
				Id:    "foo",
				Type:  github.GithubTeamResourceType,
				Attrs: &resource.Attributes{},
			},
			want: "Id: foo",
		},
		{
			name: "test with name github_team",
			res: &resource.AbstractResource{
				Id:   "foo",
				Type: github.GithubTeamResourceType,
				Attrs: &resource.Attributes{
					"name": "my-org-name",
				},
			},
			want: "Id: foo, Name: my-org-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testresource.InitFakeSchemaRepository("github", "4.4.0")
			github.InitResourcesMetadata(repo)

			abstractResource, ok := tt.res.(*resource.AbstractResource)
			if ok {
				schema, _ := repo.GetSchema(abstractResource.TerraformType())
				abstractResource.Sch = schema
			}

			if got := HumanizeAttribute(tt.res); got != tt.want {
				t.Errorf("HumanizeAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

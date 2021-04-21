package pkg_test

import (
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/jmespath/go-jmespath"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/mock"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	filter2 "github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	testresource "github.com/cloudskiff/driftctl/test/resource"
)

type TestCase struct {
	name            string
	stateResources  []resource.Resource
	remoteResources []resource.Resource
	filter          string
	mocks           func(factory resource.ResourceFactory)
	assert          func(result *test.ScanResult, err error)
}

type TestCases []TestCase

func runTest(t *testing.T, cases TestCases) {
	testresource.InitFakeResourceMetadata()
	aws.InitResourcesMetadata()
	github.InitMetadatas()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testAlerter := alerter.NewAlerter()

			if c.stateResources == nil {
				c.stateResources = []resource.Resource{}
			}
			for _, res := range c.stateResources {
				res, ok := res.(*testresource.FakeResource)
				if ok {
					impliedType, _ := gocty.ImpliedType(res)
					value, _ := gocty.ToCtyValue(res, impliedType)
					res.CtyVal = &value
					continue
				}
			}

			stateSupplier := &resource.MockSupplier{}
			stateSupplier.On("Resources").Return(c.stateResources, nil)

			if c.remoteResources == nil {
				c.remoteResources = []resource.Resource{}
			}
			for _, res := range c.remoteResources {
				res, ok := res.(*testresource.FakeResource)
				if ok {
					impliedType, _ := gocty.ImpliedType(res)
					value, _ := gocty.ToCtyValue(res, impliedType)
					res.CtyVal = &value
					continue
				}
			}

			remoteSupplier := &resource.MockSupplier{}
			remoteSupplier.On("Resources").Return(c.remoteResources, nil)

			var filter *jmespath.JMESPath
			if c.filter != "" {
				f, err := filter2.BuildExpression(c.filter)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", c.filter, err)
				}
				filter = f
			}

			resourceFactory := &terraform.MockResourceFactory{}

			if c.mocks != nil {
				c.mocks(resourceFactory)
			}

			driftctl := pkg.NewDriftCTL(remoteSupplier, stateSupplier, filter, testAlerter, resourceFactory)

			analysis, err := driftctl.Run()

			c.assert(test.NewScanResult(t, analysis), err)
		})
	}
}

func matchByAttributes(input, attrs map[string]interface{}) bool {
	for k, v := range attrs {
		if value, ok := input[k]; !ok || !reflect.DeepEqual(value, v) {
			return false
		}
	}
	return true
}

func TestDriftctlRun_BasicBehavior(t *testing.T) {

	cases := TestCases{
		{
			name: "infrastructure should be in sync",
			stateResources: []resource.Resource{
				&testresource.FakeResource{},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertInfrastructureIsInSync()
			},
		},
		{
			name: "we should have deleted resource",
			stateResources: []resource.Resource{
				&testresource.FakeResource{},
			},
			remoteResources: []resource.Resource{},
			assert: func(result *test.ScanResult, err error) {
				result.AssertDeletedCount(1)
			},
		},
		{
			name:           "we should have unmanaged resource",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
			},
		},
		{
			name: "we should have changes of field update",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:     "fake",
					FooBar: "barfoo",
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:     "fake",
					FooBar: "foobar",
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"foo_bar"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: false,
				})
			},
		},
		{
			name: "we should have changes on computed field",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:     "fake",
					BarFoo: "barfoo",
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:     "fake",
					BarFoo: "foobar",
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"bar_foo"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: true,
				})
			},
		},
		{
			name: "we should have changes of deleted field",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Tags: map[string]string{
						"tag1": "deleted",
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:   "fake",
					Tags: map[string]string{},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.DELETE,
						Path: []string{"tags", "tag1"},
						From: "deleted",
						To:   nil,
					},
					Computed: false,
				})
			},
		},
		{
			name: "we should have changes of added field",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:   "fake",
					Tags: map[string]string{},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Tags: map[string]string{
						"tag1": "added",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.CREATE,
						Path: []string{"tags", "tag1"},
						From: nil,
						To:   "added",
					},
					Computed: false,
				})
			},
		},
	}

	runTest(t, cases)
}

func TestDriftctlRun_BasicFilter(t *testing.T) {
	cases := TestCases{
		{
			name:           "test filtering on Type",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "not-filtered",
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "filtered",
				},
			},
			filter: "Type=='filtered'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
		},
		{
			name:           "test filtering on Id",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "not-filtered",
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "filtered",
				},
			},
			filter: "Id=='res2'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
		},
		{
			name:           "test filtering on attribute",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "filtered",
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"test_field": cty.StringVal("value to filter on"),
						})
						return &v
					}(),
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "not-filtered",
				},
			},
			filter: "Attr.test_field=='value to filter on'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res1", "filtered")
			},
		},
	}

	runTest(t, cases)
}

func TestDriftctlRun_Middlewares(t *testing.T) {
	cases := TestCases{
		{
			name: "test bucket policy expander middleware",
			stateResources: []resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"foo\"}"),
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"bar\"}"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":     cty.StringVal("foo"),
							"bucket": cty.StringVal("foo"),
							"policy": cty.StringVal("{\"Id\":\"bar\"}"),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"id":     cty.StringVal("foo"),
					"bucket": cty.StringVal("foo"),
					"policy": cty.StringVal("{\"Id\":\"foo\"}"),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":     "foo",
						"bucket": awssdk.String("foo"),
						"policy": awssdk.String("{\"Id\":\"foo\"}"),
					})
				}), "aws_s3_bucket_policy").Times(1).Return(&foo, nil)
			},
			filter: "Type=='aws_s3_bucket_policy' && Attr.bucket=='foo'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_s3_bucket_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"Id\":\"foo\"}",
						To:   "{\"Id\":\"bar\"}",
					},
					Computed: false,
				})
			},
		},
		{
			name: "test instance block device middleware",
			stateResources: []resource.Resource{
				&aws.AwsInstance{
					Id:               "dummy-instance",
					AvailabilityZone: awssdk.String("us-east-1"),
					EbsBlockDevice: &[]struct {
						DeleteOnTermination *bool   `cty:"delete_on_termination"`
						DeviceName          *string `cty:"device_name"`
						Encrypted           *bool   `cty:"encrypted" computed:"true"`
						Iops                *int    `cty:"iops" computed:"true"`
						KmsKeyId            *string `cty:"kms_key_id" computed:"true"`
						SnapshotId          *string `cty:"snapshot_id" computed:"true"`
						VolumeId            *string `cty:"volume_id" computed:"true"`
						VolumeSize          *int    `cty:"volume_size" computed:"true"`
						VolumeType          *string `cty:"volume_type" computed:"true"`
					}{
						{
							VolumeId:  awssdk.String("vol-018c5ae89895aca4c"),
							Encrypted: awssdk.Bool(true),
						},
					},
					RootBlockDevice: &[]struct {
						DeleteOnTermination *bool   `cty:"delete_on_termination"`
						DeviceName          *string `cty:"device_name" computed:"true"`
						Encrypted           *bool   `cty:"encrypted" computed:"true"`
						Iops                *int    `cty:"iops" computed:"true"`
						KmsKeyId            *string `cty:"kms_key_id" computed:"true"`
						VolumeId            *string `cty:"volume_id" computed:"true"`
						VolumeSize          *int    `cty:"volume_size" computed:"true"`
						VolumeType          *string `cty:"volume_type" computed:"true"`
					}{
						{
							VolumeId:   awssdk.String("vol-02862d9b39045a3a4"),
							VolumeType: awssdk.String("gp2"),
						},
					},
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsEbsVolume{
					Id:                 "vol-018c5ae89895aca4c",
					Encrypted:          awssdk.Bool(false),
					MultiAttachEnabled: awssdk.Bool(false),
					AvailabilityZone:   awssdk.String("us-east-1"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                   cty.StringVal("vol-018c5ae89895aca4c"),
							"availability_zone":    cty.StringVal("us-east-1"),
							"encrypted":            cty.BoolVal(false),
							"multi_attach_enabled": cty.BoolVal(false),
						})
						return &v
					}(),
				},
				&aws.AwsEbsVolume{
					Id:                 "vol-02862d9b39045a3a4",
					Type:               awssdk.String("gp3"),
					MultiAttachEnabled: awssdk.Bool(false),
					AvailabilityZone:   awssdk.String("us-east-1"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                   cty.StringVal("vol-02862d9b39045a3a4"),
							"availability_zone":    cty.StringVal("us-east-1"),
							"type":                 cty.StringVal("gp3"),
							"multi_attach_enabled": cty.BoolVal(false),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"id":                   cty.StringVal("vol-018c5ae89895aca4c"),
					"availability_zone":    cty.StringVal("us-east-1"),
					"encrypted":            cty.BoolVal(true),
					"multi_attach_enabled": cty.BoolVal(false),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                   "vol-018c5ae89895aca4c",
						"availability_zone":    awssdk.String("us-east-1"),
						"encrypted":            awssdk.Bool(true),
						"multi_attach_enabled": false,
					})
				}), "aws_ebs_volume").Times(1).Return(&foo, nil)

				bar := cty.ObjectVal(map[string]cty.Value{
					"id":                   cty.StringVal("vol-02862d9b39045a3a4"),
					"availability_zone":    cty.StringVal("us-east-1"),
					"type":                 cty.StringVal("gp2"),
					"multi_attach_enabled": cty.BoolVal(false),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                   "vol-02862d9b39045a3a4",
						"availability_zone":    awssdk.String("us-east-1"),
						"type":                 awssdk.String("gp2"),
						"multi_attach_enabled": false,
					})
				}), "aws_ebs_volume").Times(1).Return(&bar, nil)
			},
			filter: "Type=='aws_ebs_volume' && Attr.availability_zone=='us-east-1'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertResourceHasDrift("vol-02862d9b39045a3a4", "aws_ebs_volume", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"type"},
						From: "gp2",
						To:   "gp3",
					},
				})
				result.AssertResourceHasDrift("vol-018c5ae89895aca4c", "aws_ebs_volume", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"encrypted"},
						From: true,
						To:   false,
					},
				})
			},
		},
		{
			name: "test route table expander middleware",
			stateResources: []resource.Resource{
				&aws.AwsRouteTable{
					Id: "table",
					Route: &[]struct {
						CidrBlock              *string `cty:"cidr_block"`
						EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
						GatewayId              *string `cty:"gateway_id"`
						InstanceId             *string `cty:"instance_id"`
						Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
						LocalGatewayId         *string `cty:"local_gateway_id"`
						NatGatewayId           *string `cty:"nat_gateway_id"`
						NetworkInterfaceId     *string `cty:"network_interface_id"`
						TransitGatewayId       *string `cty:"transit_gateway_id"`
						VpcEndpointId          *string `cty:"vpc_endpoint_id"`
						VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
					}{
						{
							GatewayId: awssdk.String("igw-07b7844a8fd17a638"),
							CidrBlock: awssdk.String("0.0.0.0/0"),
						},
						{
							GatewayId:     awssdk.String("igw-07b7844a8fd17a638"),
							Ipv6CidrBlock: awssdk.String("::/0"),
						},
					},
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsRoute{
					Id:                      "r-table1080289494",
					RouteTableId:            awssdk.String("table"),
					DestinationCidrBlock:    awssdk.String("0.0.0.0/0"),
					GatewayId:               awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                  awssdk.String("CreateRoute"),
					State:                   awssdk.String("active"),
					DestinationPrefixListId: awssdk.String(""),
					InstanceOwnerId:         awssdk.String(""),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                     cty.StringVal("r-table1080289494"),
							"route_table_id":         cty.StringVal("table"),
							"destination_cidr_block": cty.StringVal("0.0.0.0/0"),
							"gateway_id":             cty.StringVal("igw-07b7844a8fd17a638"),
							"origin":                 cty.StringVal("CreateRoute"),
							"state":                  cty.StringVal("active"),
						})
						return &v
					}(),
				},
				&aws.AwsRoute{
					Id:                       "r-table2750132062",
					RouteTableId:             awssdk.String("table"),
					DestinationIpv6CidrBlock: awssdk.String("::/0"),
					GatewayId:                awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                   awssdk.String("CreateRoute"),
					State:                    awssdk.String("active"),
					DestinationPrefixListId:  awssdk.String(""),
					InstanceOwnerId:          awssdk.String(""),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                          cty.StringVal("r-table2750132062"),
							"route_table_id":              cty.StringVal("table"),
							"destination_ipv6_cidr_block": cty.StringVal("::/0"),
							"gateway_id":                  cty.StringVal("igw-07b7844a8fd17a638"),
							"origin":                      cty.StringVal("CreateRoute"),
							"state":                       cty.StringVal("active"),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"id":                     cty.StringVal("r-table1080289494"),
					"route_table_id":         cty.StringVal("table"),
					"destination_cidr_block": cty.StringVal("0.0.0.0/0"),
					"gateway_id":             cty.StringVal("igw-07b7844a8fd17a638"),
					"origin":                 cty.StringVal("CreateRoute"),
					"state":                  cty.StringVal("active"),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                     "r-table1080289494",
						"destination_cidr_block": awssdk.String("0.0.0.0/0"),
						"gateway_id":             awssdk.String("igw-07b7844a8fd17a638"),
						"origin":                 "CreateRoute",
						"route_table_id":         "table",
						"state":                  "active",
					})
				}), "aws_route").Times(1).Return(&foo, nil)

				bar := cty.ObjectVal(map[string]cty.Value{
					"id":                          cty.StringVal("r-table2750132062"),
					"route_table_id":              cty.StringVal("table"),
					"destination_ipv6_cidr_block": cty.StringVal("::/0"),
					"gateway_id":                  cty.StringVal("igw-07b7844a8fd17a638"),
					"origin":                      cty.StringVal("CreateRoute"),
					"state":                       cty.StringVal("active"),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                          "r-table2750132062",
						"destination_ipv6_cidr_block": awssdk.String("::/0"),
						"gateway_id":                  awssdk.String("igw-07b7844a8fd17a638"),
						"origin":                      "CreateRoute",
						"route_table_id":              "table",
						"state":                       "active",
					})
				}), "aws_route").Times(1).Return(&bar, nil)
			},
			filter: "Type=='aws_route' && Attr.gateway_id=='igw-07b7844a8fd17a638'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertInfrastructureIsInSync()
			},
		},
		{
			name: "test sns topic policy expander middleware",
			stateResources: []resource.Resource{
				&aws.AwsSnsTopic{
					Id:     "foo",
					Arn:    awssdk.String("arn"),
					Policy: awssdk.String("{\"policy\":\"bar\"}"),
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsSnsTopicPolicy{
					Id:     "foo",
					Arn:    awssdk.String("arn"),
					Policy: awssdk.String("{\"policy\":\"baz\"}"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":     cty.StringVal("foo"),
							"arn":    cty.StringVal("arn"),
							"policy": cty.StringVal("{\"policy\":\"baz\"}"),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"id":     cty.StringVal("foo"),
					"arn":    cty.StringVal("arn"),
					"policy": cty.StringVal("{\"policy\":\"bar\"}"),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":     "foo",
						"arn":    awssdk.String("arn"),
						"policy": awssdk.String("{\"policy\":\"bar\"}"),
					})
				}), "aws_sns_topic_policy").Times(1).Return(&foo, nil)
			},
			filter: "Type=='aws_sns_topic_policy' && Attr.arn=='arn'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_sns_topic_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"policy\":\"bar\"}",
						To:   "{\"policy\":\"baz\"}",
					},
					Computed: false,
				})
			},
		},
		{
			name: "test sqs queue policy expander middleware",
			stateResources: []resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: awssdk.String("{\"policy\":\"bar\"}"),
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"policy\":\"baz\"}"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":        cty.StringVal("foo"),
							"queue_url": cty.StringVal("foo"),
							"policy":    cty.StringVal("{\"policy\":\"baz\"}"),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"id":        cty.StringVal("foo"),
					"queue_url": cty.StringVal("foo"),
					"policy":    cty.StringVal("{\"policy\":\"bar\"}"),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":        "foo",
						"queue_url": "foo",
						"policy":    awssdk.String("{\"policy\":\"bar\"}"),
					})
				}), "aws_sqs_queue_policy").Times(1).Return(&foo, nil)
			},
			filter: "Type=='aws_sqs_queue_policy' && Attr.queue_url=='foo'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_sqs_queue_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"policy\":\"bar\"}",
						To:   "{\"policy\":\"baz\"}",
					},
					Computed: false,
				})
			},
		},
		{
			name: "test security group rule sanitizer middleware",
			stateResources: []resource.Resource{
				&aws.AwsSecurityGroupRule{
					Id:                    "sgrule-3970541193",
					Type:                  awssdk.String("ingress"),
					SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
					Protocol:              awssdk.String("tcp"),
					FromPort:              awssdk.Int(0),
					ToPort:                awssdk.Int(65535),
					Self:                  awssdk.Bool(true),
					SourceSecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                       cty.StringVal("sgrule-3970541193"),
							"type":                     cty.StringVal("ingress"),
							"security_group_id":        cty.StringVal("sg-0254c038e32f25530"),
							"protocol":                 cty.StringVal("tcp"),
							"from_port":                cty.NumberIntVal(0),
							"to_port":                  cty.NumberIntVal(65535),
							"self":                     cty.BoolVal(true),
							"source_security_group_id": cty.StringVal("sg-0254c038e32f25530"),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-845917806",
					Type:            awssdk.String("egress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{"0.0.0.0/0"},
					Ipv6CidrBlocks:  &[]string{"::/0"},
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-294318973",
					Type:            awssdk.String("ingress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{"1.2.0.0/16", "5.6.7.0/24"},
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-2471889226",
					Type:            awssdk.String("ingress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("tcp"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					PrefixListIds:   &[]string{"pl-abb451c2"},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-2471889226"),
							"type":              cty.StringVal("ingress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("tcp"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"prefix_list_ids": cty.SetVal([]cty.Value{
								cty.StringVal("pl-abb451c2"),
							}),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:                    "sgrule-3587309474",
					Type:                  awssdk.String("ingress"),
					SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
					Protocol:              awssdk.String("tcp"),
					FromPort:              awssdk.Int(0),
					ToPort:                awssdk.Int(65535),
					SourceSecurityGroupId: awssdk.String("sg-9e0204ff"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                       cty.StringVal("sgrule-3587309474"),
							"type":                     cty.StringVal("ingress"),
							"security_group_id":        cty.StringVal("sg-0254c038e32f25530"),
							"protocol":                 cty.StringVal("tcp"),
							"from_port":                cty.NumberIntVal(0),
							"to_port":                  cty.NumberIntVal(65535),
							"source_security_group_id": cty.StringVal("sg-9e0204ff"),
						})
						return &v
					}(),
				},
			},
			remoteResources: []resource.Resource{
				&aws.AwsSecurityGroupRule{
					Id:                    "sgrule-3970541193",
					Type:                  awssdk.String("ingress"),
					SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
					Protocol:              awssdk.String("tcp"),
					FromPort:              awssdk.Int(0),
					ToPort:                awssdk.Int(65535),
					Self:                  awssdk.Bool(true),
					SourceSecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                       cty.StringVal("sgrule-3970541193"),
							"type":                     cty.StringVal("ingress"),
							"security_group_id":        cty.StringVal("sg-0254c038e32f25530"),
							"protocol":                 cty.StringVal("tcp"),
							"from_port":                cty.NumberIntVal(0),
							"to_port":                  cty.NumberIntVal(65535),
							"self":                     cty.BoolVal(true),
							"source_security_group_id": cty.StringVal("sg-0254c038e32f25530"),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-1707973622",
					Type:            awssdk.String("egress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{"0.0.0.0/0"},
					Ipv6CidrBlocks:  &[]string{},
					PrefixListIds:   &[]string{},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-1707973622"),
							"type":              cty.StringVal("egress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("-1"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"cidr_blocks": cty.SetVal([]cty.Value{
								cty.StringVal("0.0.0.0/0"),
							}),
							"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
							"prefix_list_ids":  cty.SetValEmpty(cty.String),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-2821752134",
					Type:            awssdk.String("egress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{},
					Ipv6CidrBlocks:  &[]string{"::/0"},
					PrefixListIds:   &[]string{},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-2821752134"),
							"type":              cty.StringVal("egress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("-1"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"cidr_blocks":       cty.SetValEmpty(cty.String),
							"ipv6_cidr_blocks": cty.SetVal([]cty.Value{
								cty.StringVal("::/0"),
							}),
							"prefix_list_ids": cty.SetValEmpty(cty.String),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-2165103420",
					Type:            awssdk.String("ingress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{"5.6.7.0/24"},
					Ipv6CidrBlocks:  &[]string{},
					PrefixListIds:   &[]string{},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-2165103420"),
							"type":              cty.StringVal("ingress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("-1"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"cidr_blocks": cty.SetVal([]cty.Value{
								cty.StringVal("5.6.7.0/24"),
							}),
							"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
							"prefix_list_ids":  cty.SetValEmpty(cty.String),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-2582518759",
					Type:            awssdk.String("ingress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("-1"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					CidrBlocks:      &[]string{"1.2.0.0/16"},
					Ipv6CidrBlocks:  &[]string{},
					PrefixListIds:   &[]string{},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-2582518759"),
							"type":              cty.StringVal("ingress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("-1"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"cidr_blocks": cty.SetVal([]cty.Value{
								cty.StringVal("1.2.0.0/16"),
							}),
							"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
							"prefix_list_ids":  cty.SetValEmpty(cty.String),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:              "sgrule-2471889226",
					Type:            awssdk.String("ingress"),
					SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
					Protocol:        awssdk.String("tcp"),
					FromPort:        awssdk.Int(0),
					ToPort:          awssdk.Int(0),
					PrefixListIds:   &[]string{"pl-abb451c2"},
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                cty.StringVal("sgrule-2471889226"),
							"type":              cty.StringVal("ingress"),
							"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
							"protocol":          cty.StringVal("tcp"),
							"from_port":         cty.NumberIntVal(0),
							"to_port":           cty.NumberIntVal(0),
							"prefix_list_ids": cty.SetVal([]cty.Value{
								cty.StringVal("pl-abb451c2"),
							}),
						})
						return &v
					}(),
				},
				&aws.AwsSecurityGroupRule{
					Id:                    "sgrule-3587309474",
					Type:                  awssdk.String("ingress"),
					SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
					Protocol:              awssdk.String("tcp"),
					FromPort:              awssdk.Int(0),
					ToPort:                awssdk.Int(65535),
					SourceSecurityGroupId: awssdk.String("sg-9e0204ff"),
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"id":                       cty.StringVal("sgrule-3587309474"),
							"type":                     cty.StringVal("ingress"),
							"security_group_id":        cty.StringVal("sg-0254c038e32f25530"),
							"protocol":                 cty.StringVal("tcp"),
							"from_port":                cty.NumberIntVal(0),
							"to_port":                  cty.NumberIntVal(65535),
							"source_security_group_id": cty.StringVal("sg-9e0204ff"),
						})
						return &v
					}(),
				},
			},
			mocks: func(factory resource.ResourceFactory) {
				rule1 := cty.ObjectVal(map[string]cty.Value{
					"id":                cty.StringVal("sgrule-1707973622"),
					"type":              cty.StringVal("egress"),
					"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
					"protocol":          cty.StringVal("-1"),
					"from_port":         cty.NumberIntVal(0),
					"to_port":           cty.NumberIntVal(0),
					"cidr_blocks": cty.SetVal([]cty.Value{
						cty.StringVal("0.0.0.0/0"),
					}),
					"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
					"prefix_list_ids":  cty.SetValEmpty(cty.String),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                "sgrule-1707973622",
						"type":              awssdk.String("egress"),
						"security_group_id": awssdk.String("sg-0254c038e32f25530"),
						"protocol":          awssdk.String("-1"),
						"from_port":         awssdk.Int(0),
						"to_port":           awssdk.Int(0),
						"cidr_blocks":       &[]string{"0.0.0.0/0"},
						"ipv6_cidr_blocks":  &[]string{},
						"prefix_list_ids":   &[]string{},
					})
				}), "aws_security_group_rule").Times(1).Return(&rule1, nil)

				rule2 := cty.ObjectVal(map[string]cty.Value{
					"id":                cty.StringVal("sgrule-2821752134"),
					"type":              cty.StringVal("egress"),
					"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
					"protocol":          cty.StringVal("-1"),
					"from_port":         cty.NumberIntVal(0),
					"to_port":           cty.NumberIntVal(0),
					"cidr_blocks":       cty.SetValEmpty(cty.String),
					"ipv6_cidr_blocks": cty.SetVal([]cty.Value{
						cty.StringVal("::/0"),
					}),
					"prefix_list_ids": cty.SetValEmpty(cty.String),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                "sgrule-2821752134",
						"type":              awssdk.String("egress"),
						"security_group_id": awssdk.String("sg-0254c038e32f25530"),
						"protocol":          awssdk.String("-1"),
						"from_port":         awssdk.Int(0),
						"to_port":           awssdk.Int(0),
						"cidr_blocks":       &[]string{},
						"ipv6_cidr_blocks":  &[]string{"::/0"},
						"prefix_list_ids":   &[]string{},
					})
				}), "aws_security_group_rule").Times(1).Return(&rule2, nil)

				rule3 := cty.ObjectVal(map[string]cty.Value{
					"id":                cty.StringVal("sgrule-2165103420"),
					"type":              cty.StringVal("ingress"),
					"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
					"protocol":          cty.StringVal("-1"),
					"from_port":         cty.NumberIntVal(0),
					"to_port":           cty.NumberIntVal(0),
					"cidr_blocks": cty.SetVal([]cty.Value{
						cty.StringVal("5.6.7.0/24"),
					}),
					"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
					"prefix_list_ids":  cty.SetValEmpty(cty.String),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                "sgrule-2165103420",
						"type":              awssdk.String("ingress"),
						"security_group_id": awssdk.String("sg-0254c038e32f25530"),
						"protocol":          awssdk.String("-1"),
						"from_port":         awssdk.Int(0),
						"to_port":           awssdk.Int(0),
						"cidr_blocks":       &[]string{"5.6.7.0/24"},
						"ipv6_cidr_blocks":  &[]string{},
						"prefix_list_ids":   &[]string{},
					})
				}), "aws_security_group_rule").Times(1).Return(&rule3, nil)

				rule4 := cty.ObjectVal(map[string]cty.Value{
					"id":                cty.StringVal("sgrule-2582518759"),
					"type":              cty.StringVal("ingress"),
					"security_group_id": cty.StringVal("sg-0254c038e32f25530"),
					"protocol":          cty.StringVal("-1"),
					"from_port":         cty.NumberIntVal(0),
					"to_port":           cty.NumberIntVal(0),
					"cidr_blocks": cty.SetVal([]cty.Value{
						cty.StringVal("1.2.0.0/16"),
					}),
					"ipv6_cidr_blocks": cty.SetValEmpty(cty.String),
					"prefix_list_ids":  cty.SetValEmpty(cty.String),
				})
				factory.(*terraform.MockResourceFactory).On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                "sgrule-2582518759",
						"type":              awssdk.String("ingress"),
						"security_group_id": awssdk.String("sg-0254c038e32f25530"),
						"protocol":          awssdk.String("-1"),
						"from_port":         awssdk.Int(0),
						"to_port":           awssdk.Int(0),
						"cidr_blocks":       &[]string{"1.2.0.0/16"},
						"ipv6_cidr_blocks":  &[]string{},
						"prefix_list_ids":   &[]string{},
					})
				}), "aws_security_group_rule").Times(1).Return(&rule4, nil)
			},
			filter: "Type=='aws_security_group_rule' && Attr.security_group_id=='sg-0254c038e32f25530'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(7)
				result.AssertInfrastructureIsInSync()
			},
		},
	}

	runTest(t, cases)
}

package pkg_test

import (
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/mock"
)

type TestProvider struct {
	Name    string
	Version string
}

type TestCase struct {
	name            string
	provider        *TestProvider
	stateResources  []resource.Resource
	remoteResources []resource.Resource
	mocks           func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface)
	assert          func(result *test.ScanResult, err error)
	options         *pkg.ScanOptions
}

type TestCases []TestCase

func runTest(t *testing.T, cases TestCases) {
	for _, c := range cases {
		if c.provider == nil {
			c.provider = &TestProvider{
				Name:    "aws",
				Version: "3.19.0",
			}
		}
		repo := testresource.InitFakeSchemaRepository(c.provider.Name, c.provider.Version)
		aws.InitResourcesMetadata(repo)
		github.InitResourcesMetadata(repo)
		t.Run(c.name, func(t *testing.T) {
			testAlerter := alerter.NewAlerter()

			if c.stateResources == nil {
				c.stateResources = []resource.Resource{}
			}

			for _, res := range c.stateResources {
				abstractResource, ok := res.(*resource.AbstractResource)
				if ok {
					schema, _ := repo.GetSchema(abstractResource.TerraformType())
					abstractResource.Sch = schema
				}
			}

			stateSupplier := &resource.MockSupplier{}
			stateSupplier.On("Resources").Return(c.stateResources, nil)

			if c.remoteResources == nil {
				c.remoteResources = []resource.Resource{}
			}

			for _, res := range c.remoteResources {
				abstractResource, ok := res.(*resource.AbstractResource)
				if ok {
					schema, _ := repo.GetSchema(abstractResource.TerraformType())
					abstractResource.Sch = schema
				}
			}
			remoteSupplier := &resource.MockSupplier{}
			remoteSupplier.On("Resources").Return(c.remoteResources, nil)

			resourceFactory := &terraform.MockResourceFactory{}

			if c.mocks != nil {
				c.mocks(resourceFactory, repo)
			}

			if c.options == nil {
				c.options = &pkg.ScanOptions{}
			}

			scanProgress := &output.MockProgress{}
			scanProgress.On("Start").Return().Once()
			scanProgress.On("Stop").Return().Once()

			iacProgress := &output.MockProgress{}
			iacProgress.On("Start").Return().Once()
			iacProgress.On("Stop").Return().Once()

			driftIgnore := filter.NewDriftIgnore()

			driftctl := pkg.NewDriftCTL(remoteSupplier, stateSupplier, testAlerter, resourceFactory, c.options, driftIgnore, scanProgress, iacProgress, repo)

			analysis, err := driftctl.Run()

			c.assert(test.NewScanResult(t, analysis), err)
			scanProgress.AssertExpectations(t)
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
			name:            "analysis duration is set",
			stateResources:  []resource.Resource{},
			remoteResources: []resource.Resource{},
			assert: func(result *test.ScanResult, err error) {
				result.NotZero(result.Duration)
			},
		},
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
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
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
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
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
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
		},
		{
			name: "we should have changes of field update",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"foobar"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: false,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
		},
		{
			name: "we should have changes on computed field",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "fake",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"arn": "barfoo",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "fake",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"arn": "foobar",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", aws.AwsAmiResourceType, analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"arn"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: true,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
		},
		{
			name: "we should have changes on deleted field",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"tags": map[string]string{
							"tag1": "deleted",
						},
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"tags": map[string]string{},
					},
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
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
		},
		{
			name: "we should have changes of added field",
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"tags": map[string]string{},
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id: "fake",
					Attrs: &resource.Attributes{
						"tags": map[string]string{
							"tag1": "added",
						},
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
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{}
			}(t),
		},
		{
			name: "we should ignore default AWS IAM role when strict mode is disabled",
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role-test-1-policy-test-1",
					map[string]interface{}{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				).Once().Return(&resource.AbstractResource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.AbstractResource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.AbstractResource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertUnmanagedCount(2)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{
					StrictMode: false,
				}
			}(t),
		},
		{
			name: "we should not ignore default AWS IAM role when strict mode is enabled",
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role-test-1-policy-test-1",
					map[string]interface{}{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				).Once().Return(&resource.AbstractResource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.AbstractResource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.AbstractResource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertUnmanagedCount(4)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{
					StrictMode: true,
				}
			}(t),
		},
		{
			name: "we should not ignore default AWS IAM role when strict mode is enabled and a filter is specified",
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role-test-1-policy-test-1",
					map[string]interface{}{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				).Once().Return(&resource.AbstractResource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "fake",
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.AbstractResource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.AbstractResource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.AbstractResource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertCoverage(0)
				result.AssertInfrastructureIsNotSync()
				result.AssertManagedCount(0)
				result.AssertUnmanagedCount(1)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Id=='role-test-1'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{
					Filter:     f,
					StrictMode: true,
				}
			}(t),
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
				&testresource.FakeResource{
					Id:    "res1",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
				&testresource.FakeResource{
					Id:    "res2",
					Type:  "filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='filtered'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name:           "test filtering on Id",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:    "res1",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
				&testresource.FakeResource{
					Id:    "res2",
					Type:  "filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Id=='res2'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name:           "test filtering on attribute",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				&testresource.FakeResource{
					Id:   "res1",
					Type: "filtered",
					Attrs: &resource.Attributes{
						"test_field": "value to filter on",
					},
				},
				&testresource.FakeResource{
					Id:    "res2",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res1", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Attr.test_field=='value to filter on'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
	}

	runTest(t, cases)
}

func TestDriftctlRun_Middlewares(t *testing.T) {
	cases := TestCases{
		{
			name: "test bucket policy expander middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "{\"Id\":\"foo\"}",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"bucket": "foo",
						"policy": "{\"Id\":\"bar\"}",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On(
					"CreateAbstractResource",
					aws.AwsS3BucketPolicyResourceType,
					"foo",
					map[string]interface{}{
						"id":     "foo",
						"bucket": "foo",
						"policy": "{\"Id\":\"foo\"}",
					},
				).Once().Return(&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"bucket": "foo",
						"policy": "{\"Id\":\"foo\"}",
					},
					Sch: getSchema(repo, aws.AwsS3BucketPolicyResourceType),
				})
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_s3_bucket_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"Id\":\"foo\"}",
						To:   "{\"Id\":\"bar\"}",
					},
					Computed:   false,
					JsonString: true,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_s3_bucket_policy' && Attr.bucket=='foo'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test instance block device middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "dummy-instance",
					Type: "aws_instance",
					Attrs: &resource.Attributes{
						"availability_zone": "us-east-1",
						"root_block_device": []interface{}{
							map[string]interface{}{
								"volume_id":   "vol-02862d9b39045a3a4",
								"volume_type": "gp2",
							},
						},
						"ebs_block_device": []interface{}{
							map[string]interface{}{
								"volume_id": "vol-018c5ae89895aca4c",
								"encrypted": true,
							},
						},
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "vol-018c5ae89895aca4c",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"encrypted":            false,
						"multi_attach_enabled": false,
						"availability_zone":    "us-east-1",
					},
				},
				&resource.AbstractResource{
					Id:   "vol-02862d9b39045a3a4",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"type":                 "gp3",
						"multi_attach_enabled": false,
						"availability_zone":    "us-east-1",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				foo := resource.AbstractResource{
					Id:   "vol-018c5ae89895aca4c",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"encrypted":            true,
						"multi_attach_enabled": false,
						"availability_zone":    "us-east-1",
					},
					Sch: getSchema(repo, "aws_ebs_volume"),
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                   "vol-018c5ae89895aca4c",
						"availability_zone":    "us-east-1",
						"encrypted":            true,
						"multi_attach_enabled": false,
					})
				})).Times(1).Return(&foo, nil)

				bar := resource.AbstractResource{
					Id:   "vol-02862d9b39045a3a4",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"type":                 "gp2",
						"multi_attach_enabled": false,
						"availability_zone":    "us-east-1",
					},
					Sch: getSchema(repo, "aws_ebs_volume"),
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"id":                   "vol-02862d9b39045a3a4",
						"availability_zone":    "us-east-1",
						"type":                 "gp2",
						"multi_attach_enabled": false,
					})
				})).Times(1).Return(&bar, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertResourceHasDrift("vol-02862d9b39045a3a4", "aws_ebs_volume", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"type"},
						From: "gp2",
						To:   "gp3",
					},
					Computed: true,
				})
				result.AssertResourceHasDrift("vol-018c5ae89895aca4c", "aws_ebs_volume", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"encrypted"},
						From: true,
						To:   false,
					},
					Computed: true,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_ebs_volume' && Attr.availability_zone=='us-east-1'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test route table expander middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "table",
					Type: "aws_route_table",
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"gateway_id": "igw-07b7844a8fd17a638",
								"cidr_block": "0.0.0.0/0",
							},
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "",
								"ipv6_cidr_block": "::/0",
							},
						},
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "r-table1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "table",
						"origin":                 "CreateRoute",
						"destination_cidr_block": "0.0.0.0/0",
						"gateway_id":             "igw-07b7844a8fd17a638",
						"state":                  "active",
					},
				},
				&resource.AbstractResource{
					Id:   "r-table2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_route", "r-table1080289494", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"destination_cidr_block": "0.0.0.0/0",
						"gateway_id":             "igw-07b7844a8fd17a638",
						"origin":                 "CreateRoute",
						"route_table_id":         "table",
						"state":                  "active",
					})
				})).Times(1).Return(&resource.AbstractResource{
					Id:   "r-table1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "table",
						"origin":                 "CreateRoute",
						"destination_cidr_block": "0.0.0.0/0",
						"gateway_id":             "igw-07b7844a8fd17a638",
						"state":                  "active",
					},
				}, nil)
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_route", "r-table2750132062", mock.MatchedBy(func(input map[string]interface{}) bool {
					return matchByAttributes(input, map[string]interface{}{
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"origin":                      "CreateRoute",
						"route_table_id":              "table",
						"state":                       "active",
					})
				})).Times(1).Return(&resource.AbstractResource{
					Id:   "r-table2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
					},
				}, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_route' && Attr.gateway_id=='igw-07b7844a8fd17a638'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test sns topic policy expander middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "foo",
						"policy": "{\"policy\":\"bar\"}",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"arn":    "arn",
						"policy": "{\"policy\":\"baz\"}",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_sns_topic_policy", "foo", map[string]interface{}{
					"id":     "foo",
					"arn":    "arn",
					"policy": "{\"policy\":\"bar\"}",
				}).Times(1).Return(&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"arn":    "arn",
						"policy": "{\"policy\":\"bar\"}",
					},
					Sch: getSchema(repo, aws.AwsSnsTopicPolicyResourceType),
				}, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_sns_topic_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"policy\":\"bar\"}",
						To:   "{\"policy\":\"baz\"}",
					},
					Computed:   false,
					JsonString: true,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_sns_topic_policy' && Attr.arn=='arn'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test sqs queue policy expander middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSqsQueueResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"policy": "{\"policy\":\"bar\"}",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"id":        "foo",
						"queue_url": "foo",
						"policy":    "{\"policy\":\"baz\"}",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_sqs_queue_policy", "foo", map[string]interface{}{
					"id":        "foo",
					"queue_url": "foo",
					"policy":    "{\"policy\":\"bar\"}",
				}).Times(1).Return(&resource.AbstractResource{
					Id:   "foo",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"id":        "foo",
						"queue_url": "foo",
						"policy":    "{\"policy\":\"bar\"}",
					},
					Sch: getSchema(repo, aws.AwsSqsQueuePolicyResourceType),
				}, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("foo", "aws_sqs_queue_policy", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"policy"},
						From: "{\"policy\":\"bar\"}",
						To:   "{\"policy\":\"baz\"}",
					},
					Computed:   false,
					JsonString: true,
				})
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_sqs_queue_policy' && Attr.queue_url=='foo'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test security group rule sanitizer middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3970541193",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3970541193",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"self":                     true,
						"source_security_group_id": "sg-0254c038e32f25530",
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-845917806",
					Attrs: &resource.Attributes{
						"id":                "sgrule-845917806",
						"type":              "egress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-294318973",
					Attrs: &resource.Attributes{
						"id":                "sgrule-294318973",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{"1.2.0.0/16", "5.6.7.0/24"},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2471889226",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2471889226",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "tcp",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"prefix_list_ids":   []interface{}{"pl-abb451c2"},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3587309474",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3587309474",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"source_security_group_id": "sg-9e0204ff",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3970541193",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3970541193",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"self":                     true,
						"source_security_group_id": "sg-0254c038e32f25530",
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-1707973622",
					Attrs: &resource.Attributes{
						"id":                "sgrule-1707973622",
						"type":              "egress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
						"ipv6_cidr_blocks":  []interface{}{},
						"prefix_list_ids":   []interface{}{},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2821752134",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2821752134",
						"type":              "egress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{},
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
						"prefix_list_ids":   []interface{}{},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2165103420",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2165103420",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{"5.6.7.0/24"},
						"ipv6_cidr_blocks":  []interface{}{},
						"prefix_list_ids":   []interface{}{},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2582518759",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2582518759",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{"1.2.0.0/16"},
						"ipv6_cidr_blocks":  []interface{}{},
						"prefix_list_ids":   []interface{}{},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2471889226",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2471889226",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "tcp",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"prefix_list_ids":   []interface{}{"pl-abb451c2"},
					},
				},
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3587309474",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3587309474",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"source_security_group_id": "sg-9e0204ff",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				rule1 := resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-1707973622",
					Attrs: &resource.Attributes{
						"id":                "sgrule-1707973622",
						"type":              "egress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks": []interface{}{
							"0.0.0.0/0",
						},
						"ipv6_cidr_blocks": []interface{}{},
						"prefix_list_ids":  []interface{}{},
					},
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_security_group_rule", rule1.Id,
					mock.MatchedBy(func(input map[string]interface{}) bool {
						return matchByAttributes(input, map[string]interface{}{
							"id":                "sgrule-1707973622",
							"type":              "egress",
							"security_group_id": "sg-0254c038e32f25530",
							"protocol":          "-1",
							"from_port":         float64(0),
							"to_port":           float64(0),
							"cidr_blocks":       []interface{}{"0.0.0.0/0"},
							"ipv6_cidr_blocks":  []interface{}{},
							"prefix_list_ids":   []interface{}{},
						})
					})).Times(1).Return(&rule1, nil)

				rule2 := resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2821752134",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2821752134",
						"type":              "egress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks":       []interface{}{},
						"ipv6_cidr_blocks": []interface{}{
							"::/0",
						},
						"prefix_list_ids": []interface{}{},
					},
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_security_group_rule", rule2.Id,
					mock.MatchedBy(func(input map[string]interface{}) bool {
						return matchByAttributes(input, map[string]interface{}{
							"id":                "sgrule-2821752134",
							"type":              "egress",
							"security_group_id": "sg-0254c038e32f25530",
							"protocol":          "-1",
							"from_port":         float64(0),
							"to_port":           float64(0),
							"cidr_blocks":       []interface{}{},
							"ipv6_cidr_blocks":  []interface{}{"::/0"},
							"prefix_list_ids":   []interface{}{},
						})
					})).Times(1).Return(&rule2, nil)

				rule3 := resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2165103420",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2165103420",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks": []interface{}{
							"5.6.7.0/24",
						},
						"ipv6_cidr_blocks": []interface{}{},
						"prefix_list_ids":  []interface{}{},
					},
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_security_group_rule", rule3.Id,
					mock.MatchedBy(func(input map[string]interface{}) bool {
						return matchByAttributes(input, map[string]interface{}{
							"id":                "sgrule-2165103420",
							"type":              "ingress",
							"security_group_id": "sg-0254c038e32f25530",
							"protocol":          "-1",
							"from_port":         float64(0),
							"to_port":           float64(0),
							"cidr_blocks":       []interface{}{"5.6.7.0/24"},
							"ipv6_cidr_blocks":  []interface{}{},
							"prefix_list_ids":   []interface{}{},
						})
					})).Times(1).Return(&rule3, nil)

				rule4 := resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-2582518759",
					Attrs: &resource.Attributes{
						"id":                "sgrule-2582518759",
						"type":              "ingress",
						"security_group_id": "sg-0254c038e32f25530",
						"protocol":          "-1",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"cidr_blocks": []interface{}{
							"1.2.0.0/16",
						},
						"ipv6_cidr_blocks": []interface{}{},
						"prefix_list_ids":  []interface{}{},
					},
				}
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", "aws_security_group_rule", rule4.Id,
					mock.MatchedBy(func(input map[string]interface{}) bool {
						return matchByAttributes(input, map[string]interface{}{
							"id":                "sgrule-2582518759",
							"type":              "ingress",
							"security_group_id": "sg-0254c038e32f25530",
							"protocol":          "-1",
							"from_port":         float64(0),
							"to_port":           float64(0),
							"cidr_blocks":       []interface{}{"1.2.0.0/16"},
							"ipv6_cidr_blocks":  []interface{}{},
							"prefix_list_ids":   []interface{}{},
						})
					})).Times(1).Return(&rule4, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(7)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_security_group_rule' && Attr.security_group_id=='sg-0254c038e32f25530'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
		{
			name: "test iam_policy_attachment_transformer & iam_policy_attachment_expander middleware",
			stateResources: []resource.Resource{
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3970541193",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3970541193",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"self":                     true,
						"source_security_group_id": "sg-0254c038e32f25530",
					},
				},
				&resource.AbstractResource{
					Id:   "iduser1",
					Type: aws.AwsIamUserPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"user":       "user1",
					},
				},
				&resource.AbstractResource{
					Id:   "idrole1",
					Type: aws.AwsIamRolePolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"role":       "role1",
					},
				},
			},
			remoteResources: []resource.Resource{
				&resource.AbstractResource{
					Type: aws.AwsSecurityGroupRuleResourceType,
					Id:   "sgrule-3970541193",
					Attrs: &resource.Attributes{
						"id":                       "sgrule-3970541193",
						"type":                     "ingress",
						"security_group_id":        "sg-0254c038e32f25530",
						"protocol":                 "tcp",
						"from_port":                float64(0),
						"to_port":                  float64(65535),
						"self":                     true,
						"source_security_group_id": "sg-0254c038e32f25530",
					},
				},
				&resource.AbstractResource{
					Id:   "iduser1",
					Type: aws.AwsIamUserPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"user":       "user1",
					},
				},
				&resource.AbstractResource{
					Id:   "idrole1",
					Type: aws.AwsIamRolePolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"role":       "role1",
					},
				},
			},
			mocks: func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface) {
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "iduser1", map[string]interface{}{
					"id":         "iduser1",
					"policy_arn": "policy_arn1",
					"users":      []interface{}{"user1"},
					"groups":     []interface{}{},
					"roles":      []interface{}{},
				}).Twice().Return(&resource.AbstractResource{
					Id:   "id1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "iduser1",
						"policy_arn": "policy_arn1",
						"users":      []interface{}{"user1"},
						"groups":     []interface{}{},
						"roles":      []interface{}{},
					},
				}, nil)
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "user1-policy_arn1", map[string]interface{}{
					"policy_arn": "policy_arn1",
					"users":      []interface{}{"user1"},
				}).Twice().Return(&resource.AbstractResource{
					Id:   "user1-policy_arn1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"users":      []interface{}{"user1"},
					},
				}, nil)
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "idrole1", map[string]interface{}{
					"id":         "idrole1",
					"policy_arn": "policy_arn1",
					"users":      []interface{}{},
					"groups":     []interface{}{},
					"roles":      []interface{}{"role1"},
				}).Twice().Return(&resource.AbstractResource{
					Id:   "idrole1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "idrole1",
						"policy_arn": "policy_arn1",
						"users":      []interface{}{},
						"groups":     []interface{}{},
						"roles":      []interface{}{"role1"},
					},
				}, nil)
				factory.(*terraform.MockResourceFactory).On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "role1-policy_arn1", map[string]interface{}{
					"policy_arn": "policy_arn1",
					"roles":      []interface{}{"role1"},
				}).Twice().Return(&resource.AbstractResource{
					Id:   "role1-policy_arn1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"roles":      []interface{}{"role1"},
					},
				}, nil)
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_iam_policy_attachment'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f}
			}(t),
		},
	}

	runTest(t, cases)
}

func getSchema(repo resource.SchemaRepositoryInterface, resourceType string) *resource.Schema {
	sch, _ := repo.GetSchema(resourceType)
	return sch
}

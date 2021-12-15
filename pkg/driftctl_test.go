package pkg_test

import (
	"reflect"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/memstore"
	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/resource/github"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestProvider struct {
	Name    string
	Version string
}

type TestCase struct {
	name            string
	provider        *TestProvider
	stateResources  []*resource.Resource
	remoteResources []*resource.Resource
	mocks           func(factory resource.ResourceFactory, repo resource.SchemaRepositoryInterface)
	assert          func(t *testing.T, result *test.ScanResult, err error)
	assertStore     func(*testing.T, memstore.Store)
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
				c.stateResources = []*resource.Resource{}
			}

			for _, res := range c.stateResources {
				schema, _ := repo.GetSchema(res.ResourceType())
				res.Sch = schema
			}

			stateSupplier := &resource.MockSupplier{}
			stateSupplier.On("Resources").Return(c.stateResources, nil)

			if c.remoteResources == nil {
				c.remoteResources = []*resource.Resource{}
			}

			for _, res := range c.remoteResources {
				schema, _ := repo.GetSchema(res.ResourceType())
				res.Sch = schema
			}

			remoteSupplier := &resource.MockRemoteSupplier{}
			remoteSupplier.On("EnumerateResources").Return(c.remoteResources, nil)
			remoteSupplier.On("ReadResources", mock.IsType([]*resource.Resource{})).Return(c.remoteResources, nil)

			var resourceFactory resource.ResourceFactory = terraform.NewTerraformResourceFactory(repo)

			if c.mocks != nil {
				resourceFactory = &terraform.MockResourceFactory{}
				c.mocks(resourceFactory, repo)
			}

			if c.options == nil {
				c.options = &pkg.ScanOptions{Deep: true}
			}

			scanProgress := &output.MockProgress{}
			scanProgress.On("Start").Return().Times(2)
			scanProgress.On("Stop").Return().Times(2)

			iacProgress := &output.MockProgress{}
			iacProgress.On("Start").Return().Once()
			iacProgress.On("Stop").Return().Once()

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)
			testFilter.On("IsResourceIgnored", mock.Anything).Return(false)
			testFilter.On("IsFieldIgnored", mock.Anything, mock.Anything).Return(false)
			analyzer := analyser.NewAnalyzer(testAlerter, analyser.AnalyzerOptions{Deep: c.options.Deep}, testFilter)

			store := memstore.New()
			driftctl := pkg.NewDriftCTL(remoteSupplier, stateSupplier, testAlerter, analyzer, resourceFactory, c.options, scanProgress, iacProgress, repo, store)

			analysis, err := driftctl.Run()

			c.assert(t, test.NewScanResult(t, analysis), err)
			if c.assertStore != nil {
				c.assertStore(t, store)
			}
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
			stateResources:  []*resource.Resource{},
			remoteResources: []*resource.Resource{},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.NotZero(result.Duration)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 0, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 0, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name: "infrastructure should be in sync",
			stateResources: []*resource.Resource{
				&resource.Resource{},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertInfrastructureIsInSync()
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{Deep: true}
			}(t),
		},
		{
			name: "we should have deleted resource",
			stateResources: []*resource.Resource{
				&resource.Resource{},
			},
			remoteResources: []*resource.Resource{},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertDeletedCount(1)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 0, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name:           "we should have unmanaged resource",
			stateResources: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				&resource.Resource{},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 0, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name: "we should have changes of field update",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name: "we should have changes on computed field",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"arn": "barfoo",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"arn": "foobar",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name: "we should have changes on deleted field",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"tags": map[string]string{
							"tag1": "deleted",
						},
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"tags": map[string]string{},
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
		},
		{
			name: "we should have changes of added field",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"tags": map[string]string{},
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "fake",
					Type: "FakeResource",
					Attrs: &resource.Attributes{
						"tags": map[string]string{
							"tag1": "added",
						},
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
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
				).Once().Return(&resource.Resource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.Resource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.Resource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.Resource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.Resource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertUnmanagedCount(2)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 4, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 2, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{
					StrictMode: false,
					Deep:       true,
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
				).Once().Return(&resource.Resource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.Resource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.Resource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.Resource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.Resource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertUnmanagedCount(4)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 6, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 2, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				return &pkg.ScanOptions{
					StrictMode: true,
					Deep:       true,
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
				).Once().Return(&resource.Resource{
					Id:   "role-test-1-policy-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"roles":      []interface{}{"role-test-1"},
					},
				})
			},
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "fake",
					Type:  "FakeResource",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "role-test-1",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/test",
					},
				},
				&resource.Resource{
					Id:   "role-policy-test-1",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "role-test-1",
					},
				},
				&resource.Resource{
					Id:   "policy-test-1",
					Type: aws.AwsIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn": "policy-test-1",
					},
				},
				&resource.Resource{
					Id:   "policy-attachment-test-1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy-test-1",
						"users":      []interface{}{},
						"roles":      []interface{}{"role-test-1"},
					},
				},
				&resource.Resource{
					Id:   "role-test-2",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/test",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertCoverage(0)
				result.AssertInfrastructureIsNotSync()
				result.AssertManagedCount(0)
				result.AssertUnmanagedCount(1)
				result.AssertDeletedCount(0)
				result.AssertDriftCountTotal(0)
			},
			assertStore: func(t *testing.T, store memstore.Store) {
				assert.Equal(t, 1, store.Bucket(memstore.TelemetryBucket).Get("total_resources"))
				assert.Equal(t, 0, store.Bucket(memstore.TelemetryBucket).Get("total_managed"))
				assert.Equal(t, uint(0), store.Bucket(memstore.TelemetryBucket).Get("duration"))
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
					Deep:       true,
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
			stateResources: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "res1",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "res2",
					Type:  "filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='filtered'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name:           "test filtering on Id",
			stateResources: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "res1",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "res2",
					Type:  "filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Id=='res2'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name:           "test filtering on attribute",
			stateResources: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "res1",
					Type: "filtered",
					Attrs: &resource.Attributes{
						"test_field": "value to filter on",
					},
				},
				&resource.Resource{
					Id:    "res2",
					Type:  "not-filtered",
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res1", "filtered")
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Attr.test_field=='value to filter on'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
	}

	runTest(t, cases)
}

func TestDriftctlRun_Middlewares(t *testing.T) {
	cases := TestCases{
		{
			name: "test bucket policy expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "{\"Id\":\"foo\"}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				).Once().Return(&resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test instance block device middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
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
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "vol-018c5ae89895aca4c",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"encrypted":            false,
						"multi_attach_enabled": false,
						"availability_zone":    "us-east-1",
					},
				},
				&resource.Resource{
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
				foo := resource.Resource{
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

				bar := resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test route table expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
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
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				&resource.Resource{
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
				})).Times(1).Return(&resource.Resource{
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
				})).Times(1).Return(&resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_route' && Attr.gateway_id=='igw-07b7844a8fd17a638'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test sns topic policy expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "foo",
					Type: aws.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "foo",
						"policy": "{\"policy\":\"bar\"}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				}).Times(1).Return(&resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test sqs queue policy expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "foo",
					Type: aws.AwsSqsQueueResourceType,
					Attrs: &resource.Attributes{
						"id":     "foo",
						"policy": "{\"policy\":\"bar\"}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				}).Times(1).Return(&resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
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

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test security group rule sanitizer middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				&resource.Resource{
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
				rule1 := resource.Resource{
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

				rule2 := resource.Resource{
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

				rule3 := resource.Resource{
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

				rule4 := resource.Resource{
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
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertManagedCount(7)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_security_group_rule' && Attr.security_group_id=='sg-0254c038e32f25530'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test iam_policy_attachment_transformer & iam_policy_attachment_expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
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
				&resource.Resource{
					Id:   "iduser1",
					Type: aws.AwsIamUserPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"user":       "user1",
					},
				},
				&resource.Resource{
					Id:   "idrole1",
					Type: aws.AwsIamRolePolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"role":       "role1",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
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
				&resource.Resource{
					Id:   "iduser1",
					Type: aws.AwsIamUserPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"user":       "user1",
					},
				},
				&resource.Resource{
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
				}).Twice().Return(&resource.Resource{
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
				}).Twice().Return(&resource.Resource{
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
				}).Twice().Return(&resource.Resource{
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
				}).Twice().Return(&resource.Resource{
					Id:   "role1-policy_arn1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "policy_arn1",
						"roles":      []interface{}{"role1"},
					},
				}, nil)
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertManagedCount(2)
				result.AssertInfrastructureIsInSync()
			},
			options: func(t *testing.T) *pkg.ScanOptions {
				filterStr := "Type=='aws_iam_policy_attachment'"
				f, err := filter.BuildExpression(filterStr)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", filterStr, err)
				}

				return &pkg.ScanOptions{Filter: f, Deep: true}
			}(t),
		},
		{
			name: "test aws role managed policy expander",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:   "role_with_managed_policy_attr",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"name": "role_with_managed_policy_attr",
						"managed_policy_arns": []interface{}{
							"arn1",
							"arn2",
						},
					},
				},
				&resource.Resource{
					Id:   "role_with_managed_policy_attr-arn2",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "arn2",
						"roles":      []interface{}{"role_with_managed_policy_attr"},
					},
				},
				&resource.Resource{
					Id:   "role_with_empty_managed_policy_attribute",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"managed_policy_arns": []interface{}{},
					},
				},
				&resource.Resource{
					Id:    "role_without_managed_policy_attribute",
					Type:  aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:   "role_with_managed_policy_attr",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"name": "role_with_managed_policy_attr",
					},
				},
				&resource.Resource{
					Id:   "role_with_managed_policy_attr-arn1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "arn1",
						"roles":      []interface{}{"role_with_managed_policy_attr"},
					},
				},
				&resource.Resource{
					Id:   "role_with_managed_policy_attr-arn2",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"policy_arn": "arn2",
						"roles":      []interface{}{"role_with_managed_policy_attr"},
					},
				},
				&resource.Resource{
					Id:    "role_with_empty_managed_policy_attribute",
					Type:  aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "role_without_managed_policy_attribute",
					Type:  aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertInfrastructureIsInSync()
				result.AssertManagedCount(5)
			},
		},
		{
			name: "test aws eip association expander middleware",
			stateResources: []*resource.Resource{
				&resource.Resource{
					Id:    "ID",
					Type:  "ANOTHERTYPE",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "associdpresentinstate",
					Type:  aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "associdpresentinstate",
					Type: aws.AwsEipResourceType,
					Attrs: &resource.Attributes{
						"association_id": "associdpresentinstate",
					},
				},
				&resource.Resource{
					Id:   "associdNOTpresentinstate",
					Type: aws.AwsEipResourceType,
					Attrs: &resource.Attributes{
						"association_id":    "associdNOTpresentinstate",
						"instance":          "instanceidNOTpresentinstate",
						"network_interface": "networkinterface",
						"private_ip":        "privateip",
						"public_ip":         "publicip",
					},
				},
			},
			remoteResources: []*resource.Resource{
				&resource.Resource{
					Id:    "ID",
					Type:  "ANOTHERTYPE",
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "associdpresentinstate",
					Type:  aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "associdpresentinstate",
					Type: aws.AwsEipResourceType,
					Attrs: &resource.Attributes{
						"association_id": "associdpresentinstate",
					},
				},
				&resource.Resource{
					Id:   "associdNOTpresentinstate",
					Type: aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{
						"allocation_id":        "associdNOTpresentinstate",
						"id":                   "associdNOTpresentinstate",
						"instance_id":          "instanceidNOTpresentinstate",
						"network_interface_id": "networkinterface",
						"private_ip_address":   "privateip",
						"public_ip":            "publicip",
					},
				},
				&resource.Resource{
					Id:   "associdNOTpresentinstate",
					Type: aws.AwsEipResourceType,
					Attrs: &resource.Attributes{
						"association_id":    "associdNOTpresentinstate",
						"instance":          "instanceidNOTpresentinstate",
						"network_interface": "networkinterface",
						"private_ip":        "privateip",
						"public_ip":         "publicip",
					},
				},
			},
			assert: func(t *testing.T, result *test.ScanResult, err error) {
				result.AssertInfrastructureIsInSync()
				result.AssertManagedCount(5)
			},
		},
	}

	runTest(t, cases)
}

func getSchema(repo resource.SchemaRepositoryInterface, resourceType string) *resource.Schema {
	sch, _ := repo.GetSchema(resourceType)
	return sch
}

package remote

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudskiff/driftctl/mocks"
	remoteaws "github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIamUser(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam user",
			dirName: "iam_user_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return([]*iam.User{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples users",
			dirName: "iam_user_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return([]*iam.User{
					{
						UserName: aws.String("test-driftctl-0"),
					},
					{
						UserName: aws.String("test-driftctl-1"),
					},
					{
						UserName: aws.String("test-driftctl-2"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list iam user",
			dirName: "iam_user_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}
	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamUserEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamUserResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsIamUserResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamUserResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamUserPolicy(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam user policy",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicies", users).Return([]string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples users multiple policies",
			dirName: "iam_user_policy_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
					{
						UserName: aws.String("loadbalancer2"),
					},
					{
						UserName: aws.String("loadbalancer3"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicies", users).Once().Return([]string{
					*aws.String("loadbalancer:test"),
					*aws.String("loadbalancer:test2"),
					*aws.String("loadbalancer:test3"),
					*aws.String("loadbalancer:test4"),
					*aws.String("loadbalancer2:test2"),
					*aws.String("loadbalancer2:test22"),
					*aws.String("loadbalancer2:test23"),
					*aws.String("loadbalancer2:test24"),
					*aws.String("loadbalancer3:test3"),
					*aws.String("loadbalancer3:test32"),
					*aws.String("loadbalancer3:test33"),
					*aws.String("loadbalancer3:test34"),
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list user",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
		{
			test:    "cannot list user policy",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return([]*iam.User{}, nil)
				repo.On("ListAllUserPolicies", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}
	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamUserPolicyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamUserPolicyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsIamUserPolicyResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamUserPolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamPolicy(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam custom policies",
			dirName: "iam_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllPolicies").Once().Return([]*iam.Policy{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples custom policies",
			dirName: "iam_policy_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllPolicies").Once().Return([]*iam.Policy{
					{
						Arn: aws.String("arn:aws:iam::929327065333:policy/policy-0"),
					},
					{
						Arn: aws.String("arn:aws:iam::929327065333:policy/policy-1"),
					},
					{
						Arn: aws.String("arn:aws:iam::929327065333:policy/policy-2"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list iam custom policies",
			dirName: "iam_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllPolicies").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamPolicyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamPolicyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsIamPolicyResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamPolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamRole(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam roles",
			dirName: "iam_role_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples roles",
			dirName: "iam_role_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{
					{
						RoleName: aws.String("test_role_0"),
						Path:     aws.String("/"),
					},
					{
						RoleName: aws.String("test_role_1"),
						Path:     aws.String("/"),
					},
					{
						RoleName: aws.String("test_role_2"),
						Path:     aws.String("/"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam roles ignore services roles",
			dirName: "iam_role_ignore_services_roles",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{
					{
						RoleName: aws.String("AWSServiceRoleForOrganizations"),
						Path:     aws.String("/aws-service-role/organizations.amazonaws.com/"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForSupport"),
						Path:     aws.String("/aws-service-role/support.amazonaws.com/"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForTrustedAdvisor"),
						Path:     aws.String("/aws-service-role/trustedadvisor.amazonaws.com/"),
					},
				}, nil)
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamRoleEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamRoleResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsIamRoleResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamRoleResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamRolePolicyAttachment(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam role policy",
			dirName: "aws_iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test-role"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicyAttachments", roles).Return([]*repository.AttachedRolePolicy{}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples roles multiple policies",
			dirName: "aws_iam_role_policy_attachment_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test-role"),
					},
					{
						RoleName: aws.String("test-role2"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicyAttachments", roles).Return([]*repository.AttachedRolePolicy{
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy"),
							PolicyName: aws.String("test-policy"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy2"),
							PolicyName: aws.String("test-policy2"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy3"),
							PolicyName: aws.String("test-policy3"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy"),
							PolicyName: aws.String("test-policy"),
						},
						RoleName: *aws.String("test-role2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy2"),
							PolicyName: aws.String("test-policy2"),
						},
						RoleName: *aws.String("test-role2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::929327065333:policy/test-policy3"),
							PolicyName: aws.String("test-policy3"),
						},
						RoleName: *aws.String("test-role2"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples roles for ignored roles",
			dirName: "aws_iam_role_policy_attachment_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("AWSServiceRoleForSupport"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForOrganizations"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForTrustedAdvisor"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
			},
		},
		{
			test:    "Cannot list roles",
			dirName: "aws_iam_role_policy_attachment_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
		},
		{
			test:    "Cannot list roles policy attachment",
			dirName: "aws_iam_role_policy_attachment_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return([]*iam.Role{}, nil)
				repo.On("ListAllRolePolicyAttachments", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamRolePolicyAttachmentEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamRolePolicyAttachmentResourceType, remoteaws.NewIamRolePolicyAttachmentDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamRolePolicyAttachmentResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamAccessKey(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("test-driftctl"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllAccessKeys", users).Return([]*iam.AccessKeyMetadata{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples keys for multiples users",
			dirName: "iam_access_key_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("test-driftctl"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllAccessKeys", users).Return([]*iam.AccessKeyMetadata{
					{
						AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
						UserName:    aws.String("test-driftctl"),
					},
					{
						AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
						UserName:    aws.String("test-driftctl"),
					},
					{
						AccessKeyId: aws.String("AKIA5QYBVVD26EJME25D"),
						UserName:    aws.String("test-driftctl2"),
					},
					{
						AccessKeyId: aws.String("AKIA5QYBVVD2SWDFVVMG"),
						UserName:    aws.String("test-driftctl2"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list iam user",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return([]*iam.User{}, nil)
				repo.On("ListAllAccessKeys", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamAccessKeyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamAccessKeyResourceType, remoteaws.NewIamAccessKeyDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamAccessKeyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamUserPolicyAttachment(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam user policy",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicyAttachments", users).Return([]*repository.AttachedUserPolicy{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "iam multiples users multiple policies",
			dirName: "iam_user_policy_attachment_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
					{
						UserName: aws.String("loadbalancer2"),
					},
					{
						UserName: aws.String("loadbalancer3"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicyAttachments", users).Return([]*repository.AttachedUserPolicy{
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test"),
							PolicyName: aws.String("test"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test2"),
							PolicyName: aws.String("test2"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test3"),
							PolicyName: aws.String("test3"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test4"),
							PolicyName: aws.String("test4"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test"),
							PolicyName: aws.String("test"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test2"),
							PolicyName: aws.String("test2"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test3"),
							PolicyName: aws.String("test3"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test4"),
							PolicyName: aws.String("test4"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test"),
							PolicyName: aws.String("test"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test2"),
							PolicyName: aws.String("test2"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test3"),
							PolicyName: aws.String("test3"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::726421854799:policy/test4"),
							PolicyName: aws.String("test4"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
				}, nil)

			},
			wantErr: nil,
		},
		{
			test:    "cannot list user",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
		{
			test:    "cannot list user policies attachment",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return([]*iam.User{}, nil)
				repo.On("ListAllUserPolicyAttachments", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamUserPolicyAttachmentEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamUserPolicyAttachmentResourceType, remoteaws.NewIamUserPolicyAttachmentDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamUserPolicyAttachmentResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestIamRolePolicy(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		wantErr error
	}{
		{
			test:    "no iam role policy",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test_role"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicies", roles).Return([]string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiples roles with inline policies",
			dirName: "iam_role_policy_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test_role_0"),
					},
					{
						RoleName: aws.String("test_role_1"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicies", roles).Return([]string{
					*aws.String("test_role_0:policy-role0-0"),
					*aws.String("test_role_0:policy-role0-1"),
					*aws.String("test_role_0:policy-role0-2"),
					*aws.String("test_role_1:policy-role1-0"),
					*aws.String("test_role_1:policy-role1-1"),
					*aws.String("test_role_1:policy-role1-2"),
				}, nil).Once()
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list roles",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
		{
			test:    "cannot list role policy",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return([]*iam.Role{}, nil)
				repo.On("ListAllRolePolicies", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockIAMRepository{}
			c.mocks(fakeRepo)
			var repo repository.IAMRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewIAMRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(remoteaws.NewIamRolePolicyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsIamRolePolicyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsIamRolePolicyResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsIamRolePolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

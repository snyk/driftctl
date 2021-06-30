package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestVPC(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no VPC",
			dirName: "vpc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{}, []*ec2.Vpc{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "VPC results",
			dirName: "vpc",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{
					{
						VpcId: awssdk.String("vpc-0768e1fd0029e3fc3"),
					},
					{
						VpcId:     awssdk.String("vpc-020b072316a95b97f"),
						IsDefault: awssdk.Bool(false),
					},
					{
						VpcId:     awssdk.String("vpc-02c50896b59598761"),
						IsDefault: awssdk.Bool(false),
					},
				}, []*ec2.Vpc{
					{
						VpcId:     awssdk.String("vpc-a8c5d4c1"),
						IsDefault: awssdk.Bool(false),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list VPC",
			dirName: "vpc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsVpcResourceType),
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)
	alerter := &mocks.AlerterInterface{}

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockEC2Repository{}
			c.mocks(fakeRepo)
			var repo repository.EC2Repository = fakeRepo
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
				repo = repository.NewEC2Repository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsVpcResourceType, common.NewGenericDetailFetcher(resourceaws.AwsVpcResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsVpcResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestDefaultVPC(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no VPC",
			dirName: "vpc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{}, []*ec2.Vpc{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "default VPC results",
			dirName: "default_vpc",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{
					{
						VpcId:     awssdk.String("vpc-0768e1fd0029e3fc3"),
						IsDefault: awssdk.Bool(false),
					},
					{
						VpcId:     awssdk.String("vpc-020b072316a95b97f"),
						IsDefault: awssdk.Bool(false),
					},
				}, []*ec2.Vpc{
					{
						VpcId:     awssdk.String("vpc-a8c5d4c1"),
						IsDefault: awssdk.Bool(true),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list VPC",
			dirName: "vpc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllVPCs").Once().Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsDefaultVpcResourceType),
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)
	alerter := &mocks.AlerterInterface{}

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockEC2Repository{}
			c.mocks(fakeRepo)
			var repo repository.EC2Repository = fakeRepo
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
				repo = repository.NewEC2Repository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewDefaultVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultVpcResourceType, common.NewGenericDetailFetcher(resourceaws.AwsDefaultVpcResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultVpcResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

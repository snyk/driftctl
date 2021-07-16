package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
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

func TestEC2EbsVolume(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no volumes",
			dirName: "aws_ec2_ebs_volume_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{}, nil)
			},
		},
		{
			test:    "multiple volumes",
			dirName: "aws_ec2_ebs_volume_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{
					{VolumeId: awssdk.String("vol-081c7272a57a09db1")},
					{VolumeId: awssdk.String("vol-01ddc91d3d9d1318b")},
				}, nil)
			},
		},
		{
			test:    "cannot list volumes",
			dirName: "aws_ec2_ebs_volume_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2EbsVolumeEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsVolumeResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsEbsVolumeResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsVolumeResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2EbsSnapshot(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no snapshots",
			dirName: "aws_ec2_ebs_snapshot_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSnapshots").Return([]*ec2.Snapshot{}, nil)
			},
		},
		{
			test:    "multiple snapshots",
			dirName: "aws_ec2_ebs_snapshot_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSnapshots").Return([]*ec2.Snapshot{
					{SnapshotId: awssdk.String("snap-0c509a2a880d95a39")},
					{SnapshotId: awssdk.String("snap-00672558cecd93a61")},
				}, nil)
			},
		},
		{
			test:    "cannot list snapshots",
			dirName: "aws_ec2_ebs_snapshot_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSnapshots").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2EbsSnapshotEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsSnapshotResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsEbsSnapshotResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsSnapshotResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2Eip(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no eips",
			dirName: "aws_ec2_eip_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddresses").Return([]*ec2.Address{}, nil)
			},
		},
		{
			test:    "multiple eips",
			dirName: "aws_ec2_eip_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddresses").Return([]*ec2.Address{
					{AllocationId: awssdk.String("eipalloc-017d5267e4dda73f1")},
					{AllocationId: awssdk.String("eipalloc-0cf714dc097c992cc")},
				}, nil)
			},
		},
		{
			test:    "cannot list eips",
			dirName: "aws_ec2_eip_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddresses").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2EipEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEipResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsEipResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEipResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2Ami(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no amis",
			dirName: "aws_ec2_ami_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllImages").Return([]*ec2.Image{}, nil)
			},
		},
		{
			test:    "multiple amis",
			dirName: "aws_ec2_ami_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllImages").Return([]*ec2.Image{
					{ImageId: awssdk.String("ami-03a578b46f4c3081b")},
					{ImageId: awssdk.String("ami-025962fd8b456731f")},
				}, nil)
			},
		},
		{
			test:    "cannot list ami",
			dirName: "aws_ec2_ami_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllImages").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2AmiEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsAmiResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsAmiResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsAmiResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2KeyPair(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no key pairs",
			dirName: "aws_ec2_key_pair_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllKeyPairs").Return([]*ec2.KeyPairInfo{}, nil)
			},
		},
		{
			test:    "multiple key pairs",
			dirName: "aws_ec2_key_pair_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllKeyPairs").Return([]*ec2.KeyPairInfo{
					{KeyName: awssdk.String("test")},
					{KeyName: awssdk.String("bar")},
				}, nil)
			},
		},
		{
			test:    "cannot list key pairs",
			dirName: "aws_ec2_key_pair_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllKeyPairs").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2KeyPairEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsKeyPairResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsKeyPairResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsKeyPairResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2EipAssociation(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no eip associations",
			dirName: "aws_ec2_eip_association_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddressesAssociation").Return([]string{}, nil)
			},
		},
		{
			test:    "single eip association",
			dirName: "aws_ec2_eip_association_single",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddressesAssociation").Return([]string{
					"eipassoc-0e9a7356e30f0c3d1",
				}, nil)
			},
		},
		{
			test:    "cannot list eip associations",
			dirName: "aws_ec2_eip_association_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllAddressesAssociation").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2EipAssociationEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEipAssociationResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsEipAssociationResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEipAssociationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2Instance(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no instances",
			dirName: "aws_ec2_instance_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{}, nil)
			},
		},
		{
			test:    "multiple instances",
			dirName: "aws_ec2_instance_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{
					{InstanceId: awssdk.String("i-0d3650a23f4e45dc0")},
					{InstanceId: awssdk.String("i-010376047a71419f1")},
				}, nil)
			},
		},
		{
			test:    "terminated instances",
			dirName: "aws_ec2_instance_terminated",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{
					{InstanceId: awssdk.String("i-0e1543baf4f2cd990")},
					{InstanceId: awssdk.String("i-0a3a7ed51ae2b4fa0")}, // Nil
				}, nil)
			},
		},
		{
			test:    "cannot list instances",
			dirName: "aws_ec2_instance_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInstances").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2InstanceEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsInstanceResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsInstanceResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsInstanceResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2InternetGateway(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no internet gateways",
			dirName: "aws_ec2_internet_gateway_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInternetGateways").Return([]*ec2.InternetGateway{}, nil)
			},
		},
		{
			test:    "multiple internet gateways",
			dirName: "aws_ec2_internet_gateway_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInternetGateways").Return([]*ec2.InternetGateway{
					{InternetGatewayId: awssdk.String("igw-0184eb41aadc62d1c")},
					{InternetGatewayId: awssdk.String("igw-047b487f5c60fca99")},
				}, nil)
			},
		},
		{
			test:    "cannot list internet gateways",
			dirName: "aws_ec2_internet_gateway_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllInternetGateways").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2InternetGatewayEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsInternetGatewayResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsInternetGatewayResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsInternetGatewayResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

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
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsVpcResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsVpcResourceType, provider, deserializer))

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
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewDefaultVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultVpcResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsDefaultVpcResourceType, provider, deserializer))

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

func TestEC2RouteTableAssociation(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no route table associations (test for nil values)",
			dirName: "aws_ec2_route_table_association_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{
					{
						RouteTableId: awssdk.String("assoc_with_nil"),
						Associations: []*ec2.RouteTableAssociation{
							{
								AssociationState:        nil,
								GatewayId:               nil,
								Main:                    nil,
								RouteTableAssociationId: nil,
								RouteTableId:            nil,
								SubnetId:                nil,
							},
						},
					},
					{RouteTableId: awssdk.String("nil_assoc")},
				}, nil)
			},
		},
		{
			test:    "multiple route table associations (mixed subnet and gateway associations)",
			dirName: "aws_ec2_route_table_association_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{
					{
						RouteTableId: awssdk.String("rtb-05aa6c5673311a17b"), // route
						Associations: []*ec2.RouteTableAssociation{
							{ // Should be ignored
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("disassociated"),
								},
								GatewayId: awssdk.String("dummy-id"),
							},
							{ // Should be ignored
								SubnetId:  nil,
								GatewayId: nil,
							},
							{ // assoc_route_subnet1
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-0809598f92dbec03b"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-05185af647b2eeda3"),
							},
							{ // assoc_route_subnet
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-01957791b2cfe6ea4"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-0e93dbfa2e5dd8282"),
							},
							{ // assoc_route_subnet2
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								GatewayId:               nil,
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-0b4f97ea57490e213"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-0fd966efd884d0362"),
							},
						},
					},
					{
						RouteTableId: awssdk.String("rtb-09df7cc9d16de9f8f"), // route2
						Associations: []*ec2.RouteTableAssociation{
							{ // assoc_route2_gateway
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								RouteTableAssociationId: awssdk.String("rtbassoc-0a79ccacfceb4944b"),
								RouteTableId:            awssdk.String("rtb-09df7cc9d16de9f8f"),
								GatewayId:               awssdk.String("igw-0238f6e09185ac954"),
							},
						},
					},
				}, nil)
			},
		},
		{
			test:    "cannot list route table associations",
			dirName: "aws_ec2_route_table_association_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2RouteTableAssociationEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteTableAssociationResourceType, aws.NewEC2RouteTableAssociationDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteTableAssociationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2Subnet(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no subnets",
			dirName: "aws_ec2_subnet_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{}, []*ec2.Subnet{}, nil)
			},
		},
		{
			test:    "multiple subnets",
			dirName: "aws_ec2_subnet_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{
					{
						SubnetId:     awssdk.String("subnet-05810d3f933925f6d"), // subnet1
						DefaultForAz: awssdk.Bool(false),
					},
					{
						SubnetId:     awssdk.String("subnet-0b13f1e0eacf67424"), // subnet2
						DefaultForAz: awssdk.Bool(false),
					},
					{
						SubnetId:     awssdk.String("subnet-0c9b78001fe186e22"), // subnet3
						DefaultForAz: awssdk.Bool(false),
					},
				}, []*ec2.Subnet{
					{
						SubnetId:     awssdk.String("subnet-44fe0c65"), // us-east-1a
						DefaultForAz: awssdk.Bool(true),
					},
					{
						SubnetId:     awssdk.String("subnet-65e16628"), // us-east-1b
						DefaultForAz: awssdk.Bool(true),
					},
					{
						SubnetId:     awssdk.String("subnet-afa656f0"), // us-east-1c
						DefaultForAz: awssdk.Bool(true),
					},
				}, nil)
			},
		},
		{
			test:    "cannot list subnets",
			dirName: "aws_ec2_subnet_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2SubnetEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSubnetResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsSubnetResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSubnetResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2DefaultSubnet(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no default subnets",
			dirName: "aws_ec2_default_subnet_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{}, []*ec2.Subnet{}, nil)
			},
		},
		{
			test:    "multiple default subnets",
			dirName: "aws_ec2_default_subnet_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{
					{
						SubnetId:     awssdk.String("subnet-05810d3f933925f6d"), // subnet1
						DefaultForAz: awssdk.Bool(false),
					},
					{
						SubnetId:     awssdk.String("subnet-0b13f1e0eacf67424"), // subnet2
						DefaultForAz: awssdk.Bool(false),
					},
					{
						SubnetId:     awssdk.String("subnet-0c9b78001fe186e22"), // subnet3
						DefaultForAz: awssdk.Bool(false),
					},
				}, []*ec2.Subnet{
					{
						SubnetId:     awssdk.String("subnet-44fe0c65"), // us-east-1a
						DefaultForAz: awssdk.Bool(true),
					},
					{
						SubnetId:     awssdk.String("subnet-65e16628"), // us-east-1b
						DefaultForAz: awssdk.Bool(true),
					},
					{
						SubnetId:     awssdk.String("subnet-afa656f0"), // us-east-1c
						DefaultForAz: awssdk.Bool(true),
					},
				}, nil)
			},
		},
		{
			test:    "cannot list default subnets",
			dirName: "aws_ec2_default_subnet_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllSubnets").Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2DefaultSubnetEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultSubnetResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsDefaultSubnetResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultSubnetResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2RouteTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no route tables",
			dirName: "aws_ec2_route_table_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple route tables",
			dirName: "aws_ec2_route_table_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{
					{RouteTableId: awssdk.String("rtb-08b7b71af15e183ce")}, // table1
					{RouteTableId: awssdk.String("rtb-0002ac731f6fdea55")}, // table2
					{RouteTableId: awssdk.String("rtb-0c55d55593f33fbac")}, // table3
					{
						RouteTableId: awssdk.String("rtb-0eabf071c709c0976"), // default_table
						VpcId:        awssdk.String("vpc-0b4a6b3536da20ecd"),
						Associations: []*ec2.RouteTableAssociation{
							{
								Main: awssdk.Bool(true),
							},
						},
					},
				}, nil)
			},
		},
		{
			test:    "cannot list route tables",
			dirName: "aws_ec2_route_table_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2RouteTableEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteTableResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsRouteTableResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2DefaultRouteTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no default route tables",
			dirName: "aws_ec2_default_route_table_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple default route tables",
			dirName: "aws_ec2_default_route_table_single",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{
					{RouteTableId: awssdk.String("rtb-08b7b71af15e183ce")}, // table1
					{RouteTableId: awssdk.String("rtb-0002ac731f6fdea55")}, // table2
					{RouteTableId: awssdk.String("rtb-0c55d55593f33fbac")}, // table3
					{
						RouteTableId: awssdk.String("rtb-0eabf071c709c0976"), // default_table
						VpcId:        awssdk.String("vpc-0b4a6b3536da20ecd"),
						Associations: []*ec2.RouteTableAssociation{
							{
								Main: awssdk.Bool(true),
							},
						},
					},
				}, nil)
			},
		},
		{
			test:    "cannot list default route tables",
			dirName: "aws_ec2_default_route_table_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2DefaultRouteTableEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultRouteTableResourceType, aws.NewEC2DefaultRouteTableDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultRouteTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestVpcSecurityGroup(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no security groups",
			dirName: "vpc_security_group_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{}, []*ec2.SecurityGroup{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "with security groups",
			dirName: "vpc_security_group_multiple",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId:   awssdk.String("sg-0254c038e32f25530"),
						GroupName: awssdk.String("foo"),
					},
				}, []*ec2.SecurityGroup{
					{
						GroupId:   awssdk.String("sg-9e0204ff"),
						GroupName: awssdk.String("default"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list security groups",
			dirName: "vpc_security_group_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewVPCSecurityGroupEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSecurityGroupResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsSecurityGroupResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSecurityGroupResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestVpcDefaultSecurityGroup(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no security groups",
			dirName: "vpc_default_security_group_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{}, []*ec2.SecurityGroup{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "with security groups",
			dirName: "vpc_default_security_group_multiple",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId:   awssdk.String("sg-0254c038e32f25530"),
						GroupName: awssdk.String("foo"),
					},
				}, []*ec2.SecurityGroup{
					{
						GroupId:   awssdk.String("sg-9e0204ff"),
						GroupName: awssdk.String("default"),
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list security groups",
			dirName: "vpc_default_security_group_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewVPCDefaultSecurityGroupEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultSecurityGroupResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsDefaultSecurityGroupResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultSecurityGroupResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2NatGateway(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no nat gateways",
			dirName: "aws_ec2_nat_gateway_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllNatGateways").Return([]*ec2.NatGateway{}, nil)
			},
		},
		{
			test:    "single nat gateway",
			dirName: "aws_ec2_nat_gateway_single",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllNatGateways").Return([]*ec2.NatGateway{
					{NatGatewayId: awssdk.String("nat-0a5408508b19ef490")},
				}, nil)
			},
		},
		{
			test:    "cannot list nat gateways",
			dirName: "aws_ec2_nat_gateway_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllNatGateways").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2NatGatewayEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsNatGatewayResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsNatGatewayResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsNatGatewayResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestEC2Route(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			// route table with no routes case is not possible
			// as a default route will always be present in each route table
			test:    "no routes",
			dirName: "aws_ec2_route_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple routes (mixed default_route_table and route_table)",
			dirName: "aws_ec2_route_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{
					{
						RouteTableId: awssdk.String("rtb-096bdfb69309c54c3"), // table1
						Routes: []*ec2.Route{
							{
								DestinationCidrBlock: awssdk.String("10.0.0.0/16"),
								Origin:               awssdk.String("CreateRouteTable"), // default route
							},
							{
								DestinationCidrBlock: awssdk.String("1.1.1.1/32"),
								GatewayId:            awssdk.String("igw-030e74f73bd67f21b"),
								Origin:               awssdk.String("CreateRoute"),
							},
							{
								DestinationIpv6CidrBlock: awssdk.String("::/0"),
								GatewayId:                awssdk.String("igw-030e74f73bd67f21b"),
								Origin:                   awssdk.String("CreateRoute"),
							},
						},
					},
					{
						RouteTableId: awssdk.String("rtb-0169b0937fd963ddc"), // table2
						Routes: []*ec2.Route{
							{
								DestinationCidrBlock: awssdk.String("10.0.0.0/16"),
								Origin:               awssdk.String("CreateRouteTable"), // default route
							},
							{
								DestinationCidrBlock: awssdk.String("0.0.0.0/0"),
								GatewayId:            awssdk.String("igw-030e74f73bd67f21b"),
								Origin:               awssdk.String("CreateRoute"),
							},
							{
								DestinationIpv6CidrBlock: awssdk.String("::/0"),
								GatewayId:                awssdk.String("igw-030e74f73bd67f21b"),
								Origin:                   awssdk.String("CreateRoute"),
							},
						},
					},
					{
						RouteTableId: awssdk.String("rtb-02780c485f0be93c5"), // default_table
						VpcId:        awssdk.String("vpc-09fe5abc2309ba49d"),
						Associations: []*ec2.RouteTableAssociation{
							{
								Main: awssdk.Bool(true),
							},
						},
						Routes: []*ec2.Route{
							{
								DestinationCidrBlock: awssdk.String("10.0.0.0/16"),
								Origin:               awssdk.String("CreateRouteTable"), // default route
							},
							{
								DestinationCidrBlock: awssdk.String("10.1.1.0/24"),
								GatewayId:            awssdk.String("igw-030e74f73bd67f21b"),
								Origin:               awssdk.String("CreateRoute"),
							},
							{
								DestinationCidrBlock: awssdk.String("10.1.2.0/24"),
								GatewayId:            awssdk.String("igw-030e74f73bd67f21b"),
								Origin:               awssdk.String("CreateRoute"),
							},
						},
					},
					{
						RouteTableId: awssdk.String(""), // table3
						Routes: []*ec2.Route{
							{
								DestinationCidrBlock: awssdk.String("10.0.0.0/16"),
								Origin:               awssdk.String("CreateRouteTable"), // default route
							},
						},
					},
				}, nil)
			},
		},
		{
			test:    "cannot list routes",
			dirName: "aws_ec2_route_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllRouteTables").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2RouteEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteResourceType, aws.NewEC2RouteDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestVpcSecurityGroupRule(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no security group rules",
			dirName: "vpc_security_group_rule_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId:             awssdk.String("sg-0254c038e32f25530"),
						IpPermissions:       []*ec2.IpPermission{},
						IpPermissionsEgress: []*ec2.IpPermission{},
					},
				}, nil, nil)
			},
			wantErr: nil,
		},
		{
			test:    "with security group rules",
			dirName: "vpc_security_group_rule_multiple",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId: awssdk.String("sg-0254c038e32f25530"),
						IpPermissions: []*ec2.IpPermission{
							{
								FromPort:   awssdk.Int64(0),
								ToPort:     awssdk.Int64(65535),
								IpProtocol: awssdk.String("tcp"),
								UserIdGroupPairs: []*ec2.UserIdGroupPair{
									{
										GroupId: awssdk.String("sg-0254c038e32f25530"),
									},
									{
										GroupId: awssdk.String("sg-9e0204ff"),
									},
								},
							},
							{
								IpProtocol: awssdk.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: awssdk.String("1.2.0.0/16"),
									},
									{
										CidrIp: awssdk.String("5.6.7.0/24"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: awssdk.String("::/0"),
									},
								},
							},
						},
						IpPermissionsEgress: []*ec2.IpPermission{
							{
								IpProtocol: awssdk.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: awssdk.String("0.0.0.0/0"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: awssdk.String("::/0"),
									},
								},
							},
						},
					},
					{
						GroupId: awssdk.String("sg-0cc8b3c3c2851705a"),
						IpPermissions: []*ec2.IpPermission{
							{
								FromPort:   awssdk.Int64(443),
								ToPort:     awssdk.Int64(443),
								IpProtocol: awssdk.String("tcp"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: awssdk.String("0.0.0.0/0"),
									},
								},
							},
						},
						IpPermissionsEgress: []*ec2.IpPermission{
							{
								IpProtocol: awssdk.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: awssdk.String("0.0.0.0/0"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: awssdk.String("::/0"),
									},
								},
							},
							{
								IpProtocol: awssdk.String("5"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: awssdk.String("0.0.0.0/0"),
									},
								},
							},
						},
					},
				}, nil, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list security group rules",
			dirName: "vpc_security_group_rule_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

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
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
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

			remoteLibrary.AddEnumerator(aws.NewVPCSecurityGroupRuleEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSecurityGroupRuleResourceType, aws.NewVPCSecurityGroupRuleDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSecurityGroupRuleResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

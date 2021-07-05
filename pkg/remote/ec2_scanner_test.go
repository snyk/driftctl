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
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEbsVolumeResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEbsSnapshotResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEipResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsAmiResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsKeyPairResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEipAssociationResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsInstanceResourceType),
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

			sess := session.Must(session.NewSessionWithOptions(session.Options{
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
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2InstanceEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsInstanceResourceType, aws.NewEC2InstanceDetailsFetcher(provider, deserializer))

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

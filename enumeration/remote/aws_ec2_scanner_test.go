package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	aws2 "github.com/snyk/driftctl/enumeration/remote/aws"
	repository2 "github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	common2 "github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	terraform3 "github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEC2EbsVolume(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no volumes",
			dirName: "aws_ec2_ebs_volume_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{}, nil)
			},
		},
		{
			test:    "multiple volumes",
			dirName: "aws_ec2_ebs_volume_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{
					{VolumeId: awssdk.String("vol-081c7272a57a09db1")},
					{VolumeId: awssdk.String("vol-01ddc91d3d9d1318b")},
				}, nil)
			},
		},
		{
			test:    "cannot list volumes",
			dirName: "aws_ec2_ebs_volume_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllVolumes").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEbsVolumeResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEbsVolumeResourceType, resourceaws.AwsEbsVolumeResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2EbsVolumeEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsVolumeResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEbsVolumeResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsVolumeResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2EbsSnapshot(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no snapshots",
			dirName: "aws_ec2_ebs_snapshot_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSnapshots").Return([]*ec2.Snapshot{}, nil)
			},
		},
		{
			test:    "multiple snapshots",
			dirName: "aws_ec2_ebs_snapshot_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSnapshots").Return([]*ec2.Snapshot{
					{SnapshotId: awssdk.String("snap-0c509a2a880d95a39")},
					{SnapshotId: awssdk.String("snap-00672558cecd93a61")},
				}, nil)
			},
		},
		{
			test:    "cannot list snapshots",
			dirName: "aws_ec2_ebs_snapshot_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllSnapshots").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEbsSnapshotResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEbsSnapshotResourceType, resourceaws.AwsEbsSnapshotResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2EbsSnapshotEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsSnapshotResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEbsSnapshotResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsSnapshotResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2Eip(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no eips",
			dirName: "aws_ec2_eip_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAddresses").Return([]*ec2.Address{
					{}, // Test Eip without AllocationId because it can happen (seen in sentry)
				}, nil)
			},
		},
		{
			test:    "multiple eips",
			dirName: "aws_ec2_eip_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAddresses").Return([]*ec2.Address{
					{AllocationId: awssdk.String("eipalloc-017d5267e4dda73f1")},
					{AllocationId: awssdk.String("eipalloc-0cf714dc097c992cc")},
				}, nil)
			},
		},
		{
			test:    "cannot list eips",
			dirName: "aws_ec2_eip_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllAddresses").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEipResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEipResourceType, resourceaws.AwsEipResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2EipEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEipResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEipResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEipResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2Ami(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no amis",
			dirName: "aws_ec2_ami_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*ec2.Image{}, nil)
			},
		},
		{
			test:    "multiple amis",
			dirName: "aws_ec2_ami_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*ec2.Image{
					{ImageId: awssdk.String("ami-03a578b46f4c3081b")},
					{ImageId: awssdk.String("ami-025962fd8b456731f")},
				}, nil)
			},
		},
		{
			test:    "cannot list ami",
			dirName: "aws_ec2_ami_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllImages").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsAmiResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsAmiResourceType, resourceaws.AwsAmiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2AmiEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsAmiResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsAmiResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsAmiResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2KeyPair(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no key pairs",
			dirName: "aws_ec2_key_pair_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllKeyPairs").Return([]*ec2.KeyPairInfo{}, nil)
			},
		},
		{
			test:    "multiple key pairs",
			dirName: "aws_ec2_key_pair_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllKeyPairs").Return([]*ec2.KeyPairInfo{
					{KeyName: awssdk.String("test")},
					{KeyName: awssdk.String("bar")},
				}, nil)
			},
		},
		{
			test:    "cannot list key pairs",
			dirName: "aws_ec2_key_pair_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllKeyPairs").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsKeyPairResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsKeyPairResourceType, resourceaws.AwsKeyPairResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2KeyPairEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsKeyPairResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsKeyPairResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsKeyPairResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2EipAssociation(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no eip associations",
			dirName: "aws_ec2_eip_association_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAddressesAssociation").Return([]*ec2.Address{}, nil)
			},
		},
		{
			test:    "single eip association",
			dirName: "aws_ec2_eip_association_single",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAddressesAssociation").Return([]*ec2.Address{
					{
						AssociationId: awssdk.String("eipassoc-0e9a7356e30f0c3d1"),
						AllocationId:  awssdk.String("eipalloc-017d5267e4dda73f1"),
					},
				}, nil)
			},
		},
		{
			test:    "cannot list eip associations",
			dirName: "aws_ec2_eip_association_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllAddressesAssociation").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEipAssociationResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEipAssociationResourceType, resourceaws.AwsEipAssociationResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2EipAssociationEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEipAssociationResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEipAssociationResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEipAssociationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2Instance(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no instances",
			dirName: "aws_ec2_instance_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{}, nil)
			},
		},
		{
			test:    "multiple instances",
			dirName: "aws_ec2_instance_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{
					{InstanceId: awssdk.String("i-0d3650a23f4e45dc0")},
					{InstanceId: awssdk.String("i-010376047a71419f1")},
				}, nil)
			},
		},
		{
			test:    "terminated instances",
			dirName: "aws_ec2_instance_terminated",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllInstances").Return([]*ec2.Instance{
					{InstanceId: awssdk.String("i-0e1543baf4f2cd990")},
					{InstanceId: awssdk.String("i-0a3a7ed51ae2b4fa0")}, // Nil
				}, nil)
			},
		},
		{
			test:    "cannot list instances",
			dirName: "aws_ec2_instance_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllInstances").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsInstanceResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsInstanceResourceType, resourceaws.AwsInstanceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2InstanceEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsInstanceResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsInstanceResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsInstanceResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2InternetGateway(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no internet gateways",
			dirName: "aws_ec2_internet_gateway_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllInternetGateways").Return([]*ec2.InternetGateway{}, nil)
			},
		},
		{
			test:    "multiple internet gateways",
			dirName: "aws_ec2_internet_gateway_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllInternetGateways").Return([]*ec2.InternetGateway{
					{InternetGatewayId: awssdk.String("igw-0184eb41aadc62d1c")},
					{InternetGatewayId: awssdk.String("igw-047b487f5c60fca99")},
				}, nil)
			},
		},
		{
			test:    "cannot list internet gateways",
			dirName: "aws_ec2_internet_gateway_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllInternetGateways").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsInternetGatewayResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsInternetGatewayResourceType, resourceaws.AwsInternetGatewayResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2InternetGatewayEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsInternetGatewayResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsInternetGatewayResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsInternetGatewayResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestVPC(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no VPC",
			dirName: "aws_vpc_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{}, []*ec2.Vpc{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "VPC results",
			dirName: "aws_vpc",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllVPCs").Once().Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsVpcResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsVpcResourceType, resourceaws.AwsVpcResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsVpcResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsVpcResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsVpcResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestDefaultVPC(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no VPC",
			dirName: "aws_vpc_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllVPCs").Once().Return([]*ec2.Vpc{}, []*ec2.Vpc{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "default VPC results",
			dirName: "aws_default_vpc",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllVPCs").Once().Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDefaultVpcResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDefaultVpcResourceType, resourceaws.AwsDefaultVpcResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewDefaultVPCEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultVpcResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsDefaultVpcResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultVpcResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2RouteTableAssociation(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no route table associations (test for nil values)",
			dirName: "aws_ec2_route_table_association_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllRouteTables").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsRouteTableAssociationResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRouteTableAssociationResourceType, resourceaws.AwsRouteTableResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2RouteTableAssociationEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteTableAssociationResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsRouteTableAssociationResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteTableAssociationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2Subnet(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no subnets",
			dirName: "aws_ec2_subnet_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{}, []*ec2.Subnet{}, nil)
			},
		},
		{
			test:    "multiple subnets",
			dirName: "aws_ec2_subnet_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllSubnets").Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSubnetResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSubnetResourceType, resourceaws.AwsSubnetResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2SubnetEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSubnetResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsSubnetResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSubnetResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2DefaultSubnet(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no default subnets",
			dirName: "aws_ec2_default_subnet_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSubnets").Return([]*ec2.Subnet{}, []*ec2.Subnet{}, nil)
			},
		},
		{
			test:    "multiple default subnets",
			dirName: "aws_ec2_default_subnet_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllSubnets").Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDefaultSubnetResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDefaultSubnetResourceType, resourceaws.AwsDefaultSubnetResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2DefaultSubnetEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultSubnetResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsDefaultSubnetResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultSubnetResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2RouteTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no route tables",
			dirName: "aws_ec2_route_table_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple route tables",
			dirName: "aws_ec2_route_table_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllRouteTables").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsRouteTableResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRouteTableResourceType, resourceaws.AwsRouteTableResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2RouteTableEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteTableResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsRouteTableResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2DefaultRouteTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no default route tables",
			dirName: "aws_ec2_default_route_table_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple default route tables",
			dirName: "aws_ec2_default_route_table_single",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllRouteTables").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDefaultRouteTableResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDefaultRouteTableResourceType, resourceaws.AwsDefaultRouteTableResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2DefaultRouteTableEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultRouteTableResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsDefaultRouteTableResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultRouteTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestVpcSecurityGroup(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no security groups",
			dirName: "aws_vpc_security_group_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{}, []*ec2.SecurityGroup{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "with security groups",
			dirName: "aws_vpc_security_group_multiple",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_security_group_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllSecurityGroups").Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSecurityGroupResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSecurityGroupResourceType, resourceaws.AwsSecurityGroupResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewVPCSecurityGroupEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSecurityGroupResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsSecurityGroupResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSecurityGroupResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestVpcDefaultSecurityGroup(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no security groups",
			dirName: "aws_vpc_default_security_group_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{}, []*ec2.SecurityGroup{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "with security groups",
			dirName: "aws_vpc_default_security_group_multiple",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_default_security_group_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllSecurityGroups").Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDefaultSecurityGroupResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDefaultSecurityGroupResourceType, resourceaws.AwsDefaultSecurityGroupResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewVPCDefaultSecurityGroupEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultSecurityGroupResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsDefaultSecurityGroupResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultSecurityGroupResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2NatGateway(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no nat gateways",
			dirName: "aws_ec2_nat_gateway_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNatGateways").Return([]*ec2.NatGateway{}, nil)
			},
		},
		{
			test:    "single nat gateway",
			dirName: "aws_ec2_nat_gateway_single",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNatGateways").Return([]*ec2.NatGateway{
					{NatGatewayId: awssdk.String("nat-0a5408508b19ef490")},
				}, nil)
			},
		},
		{
			test:    "cannot list nat gateways",
			dirName: "aws_ec2_nat_gateway_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllNatGateways").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsNatGatewayResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsNatGatewayResourceType, resourceaws.AwsNatGatewayResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2NatGatewayEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsNatGatewayResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsNatGatewayResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsNatGatewayResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2NetworkACL(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no network ACL",
			dirName: "aws_ec2_network_acl_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{}, nil)
			},
		},
		{
			test:    "network acl",
			dirName: "aws_ec2_network_acl",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{
					{
						NetworkAclId: awssdk.String("acl-043880b4682d2366b"),
						IsDefault:    awssdk.Bool(false),
					},
					{
						NetworkAclId: awssdk.String("acl-07a565dbe518c0713"),
						IsDefault:    awssdk.Bool(false),
					},
					{
						NetworkAclId: awssdk.String("acl-e88ee595"),
						IsDefault:    awssdk.Bool(true),
					},
				}, nil)
			},
		},
		{
			test:    "cannot list network acl",
			dirName: "aws_ec2_network_acl_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllNetworkACLs").Return(nil, awsError)

				alerter.On("SendAlert",
					resourceaws.AwsNetworkACLResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common2.RemoteAWSTerraform,
						remoteerr.NewResourceListingErrorWithType(
							awsError,
							resourceaws.AwsNetworkACLResourceType,
							resourceaws.AwsNetworkACLResourceType,
						),
						alerts.EnumerationPhase,
					),
				).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2NetworkACLEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsNetworkACLResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsNetworkACLResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsNetworkACLResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2NetworkACLRule(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no network ACL",
			dirName: "aws_ec2_network_acl_rule_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{}, nil)
			},
		},
		{
			test:    "network acl rules",
			dirName: "aws_ec2_network_acl_rule",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{
					{
						NetworkAclId: awssdk.String("acl-0ad6d657494d17ee2"), // test
						IsDefault:    awssdk.Bool(false),
						Entries: []*ec2.NetworkAclEntry{
							{
								Egress:     awssdk.Bool(false),
								RuleNumber: awssdk.Int64(100),
								Protocol:   awssdk.String("6"), // tcp
								RuleAction: awssdk.String("deny"),
								CidrBlock:  awssdk.String("0.0.0.0/0"),
							},
							{
								Egress:        awssdk.Bool(false),
								RuleNumber:    awssdk.Int64(200),
								Protocol:      awssdk.String("6"), // tcp
								RuleAction:    awssdk.String("allow"),
								Ipv6CidrBlock: awssdk.String("::/0"),
							},
							{
								Egress:     awssdk.Bool(true),
								RuleNumber: awssdk.Int64(100),
								Protocol:   awssdk.String("17"), // udp
								RuleAction: awssdk.String("allow"),
								CidrBlock:  awssdk.String("172.16.1.0/0"),
							},
						},
					},
					{
						NetworkAclId: awssdk.String("acl-0de54ef59074b622e"), // test2
						IsDefault:    awssdk.Bool(false),
						Entries: []*ec2.NetworkAclEntry{
							{
								Egress:     awssdk.Bool(false),
								RuleNumber: awssdk.Int64(100),
								Protocol:   awssdk.String("17"), // udp
								RuleAction: awssdk.String("deny"),
								CidrBlock:  awssdk.String("0.0.0.0/0"),
							},
							{
								Egress:     awssdk.Bool(true),
								RuleNumber: awssdk.Int64(100),
								Protocol:   awssdk.String("17"), // udp
								RuleAction: awssdk.String("allow"),
								CidrBlock:  awssdk.String("172.16.1.0/0"),
							},
						},
					},
				}, nil)
			},
		},
		{
			test:    "cannot list network acl",
			dirName: "aws_ec2_network_acl_rule_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllNetworkACLs").Return(nil, awsError)

				alerter.On("SendAlert",
					resourceaws.AwsNetworkACLRuleResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common2.RemoteAWSTerraform,
						remoteerr.NewResourceListingErrorWithType(
							awsError,
							resourceaws.AwsNetworkACLRuleResourceType,
							resourceaws.AwsNetworkACLResourceType,
						),
						alerts.EnumerationPhase,
					),
				).Return()
			},
			wantErr: nil,
		},
	}

	version := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", version)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
			providerVersion := version
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
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2NetworkACLRuleEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsNetworkACLRuleResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsNetworkACLRuleResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsNetworkACLRuleResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2DefaultNetworkACL(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no network ACL",
			dirName: "aws_ec2_default_network_acl_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{}, nil)
			},
		},
		{
			test:    "default network acl",
			dirName: "aws_ec2_default_network_acl",
			mocks: func(repository *repository.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllNetworkACLs").Return([]*ec2.NetworkAcl{
					{
						NetworkAclId: awssdk.String("acl-043880b4682d2366b"),
						IsDefault:    awssdk.Bool(false),
					},
					{
						NetworkAclId: awssdk.String("acl-07a565dbe518c0713"),
						IsDefault:    awssdk.Bool(false),
					},
					{
						NetworkAclId: awssdk.String("acl-e88ee595"),
						IsDefault:    awssdk.Bool(true),
					},
				}, nil)
			},
		},
		{
			test:    "cannot list default network acl",
			dirName: "aws_ec2_default_network_acl_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllNetworkACLs").Return(nil, awsError)

				alerter.On("SendAlert",
					resourceaws.AwsDefaultNetworkACLResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common2.RemoteAWSTerraform,
						remoteerr.NewResourceListingErrorWithType(
							awsError,
							resourceaws.AwsDefaultNetworkACLResourceType,
							resourceaws.AwsDefaultNetworkACLResourceType,
						),
						alerts.EnumerationPhase,
					),
				).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2DefaultNetworkACLEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDefaultNetworkACLResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsDefaultNetworkACLResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDefaultNetworkACLResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2Route(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			// route table with no routes case is not possible
			// as a default route will always be present in each route table
			test:    "no routes",
			dirName: "aws_ec2_route_empty",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*ec2.RouteTable{}, nil)
			},
		},
		{
			test:    "multiple routes (mixed default_route_table and route_table)",
			dirName: "aws_ec2_route_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllRouteTables").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsRouteResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRouteResourceType, resourceaws.AwsRouteTableResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2RouteEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRouteResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsRouteResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRouteResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestVpcSecurityGroupRule(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no security group rules",
			dirName: "aws_vpc_security_group_rule_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_security_group_rule_multiple",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
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
			dirName: "aws_vpc_security_group_rule_empty",
			mocks: func(client *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllSecurityGroups").Once().Return(nil, nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSecurityGroupRuleResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSecurityGroupRuleResourceType, resourceaws.AwsSecurityGroupResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewVPCSecurityGroupRuleEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSecurityGroupRuleResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsSecurityGroupRuleResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSecurityGroupRuleResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestEC2LaunchTemplate(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no launch template",
			dirName: "aws_launch_template",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("DescribeLaunchTemplates").Return([]*ec2.LaunchTemplate{}, nil)
			},
		},
		{
			test:    "multiple launch templates",
			dirName: "aws_launch_template_multiple",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				launchTemplates := []*ec2.LaunchTemplate{
					{LaunchTemplateId: awssdk.String("lt-0ed993d09ce6afc67"), LatestVersionNumber: awssdk.Int64(1)},
					{LaunchTemplateId: awssdk.String("lt-00b2d18c6cee7fe23"), LatestVersionNumber: awssdk.Int64(1)},
				}

				repository.On("DescribeLaunchTemplates").Return(launchTemplates, nil)
			},
		},
		{
			test:    "cannot list launch templates",
			dirName: "aws_launch_template",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("DescribeLaunchTemplates").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLaunchTemplateResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLaunchTemplateResourceType, resourceaws.AwsLaunchTemplateResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewLaunchTemplateEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsLaunchTemplateResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsLaunchTemplateResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsLaunchTemplateResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestEC2EbsEncryptionByDefault(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockEC2Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no encryption by default resource",
			dirName: "aws_ebs_encryption_by_default_list",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				repository.On("IsEbsEncryptionEnabledByDefault").Return(false, nil)
			},
		},
		{
			test:    "cannot list encryption by default resources",
			dirName: "aws_ebs_encryption_by_default_error",
			mocks: func(repository *repository2.MockEC2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("IsEbsEncryptionEnabledByDefault").Return(false, awsError)

				alerter.On("SendAlert", resourceaws.AwsEbsEncryptionByDefaultResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEbsEncryptionByDefaultResourceType, resourceaws.AwsEbsEncryptionByDefaultResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockEC2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.EC2Repository = fakeRepo
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
				repo = repository2.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewEC2EbsEncryptionByDefaultEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsEncryptionByDefaultResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEbsEncryptionByDefaultResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsEncryptionByDefaultResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

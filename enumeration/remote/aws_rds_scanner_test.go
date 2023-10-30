package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRDSDBInstance(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockRDSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no db instances",
			dirName: "aws_rds_db_instance_empty",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "single db instance",
			dirName: "aws_rds_db_instance_single",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "terraform-20201015115018309600000001", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsDbInstanceResourceType, got[0].ResourceType())
			},
		},
		{
			test:    "multiple mixed db instances",
			dirName: "aws_rds_db_instance_multiple",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001")},
					{DBInstanceIdentifier: awssdk.String("database-1")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "terraform-20201015115018309600000001", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsDbInstanceResourceType, got[0].ResourceType())

				assert.Equal(t, "database-1", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsDbInstanceResourceType, got[1].ResourceType())
			},
		},
		{
			test:    "cannot list db instances",
			dirName: "aws_rds_db_instance_list",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllDBInstances").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDbInstanceResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDbInstanceResourceType, resourceaws.AwsDbInstanceResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockRDSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.RDSRepository = fakeRepo
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
				repo = repository.NewRDSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRDSDBInstanceEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestRDSDBSubnetGroup(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockRDSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no db subnet groups",
			dirName: "aws_rds_db_subnet_group_empty",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBSubnetGroups").Return([]*rds.DBSubnetGroup{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "multiple db subnet groups",
			dirName: "aws_rds_db_subnet_group_multiple",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBSubnetGroups").Return([]*rds.DBSubnetGroup{
					{DBSubnetGroupName: awssdk.String("foo")},
					{DBSubnetGroupName: awssdk.String("bar")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "foo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsDbSubnetGroupResourceType, got[0].ResourceType())

				assert.Equal(t, "bar", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsDbSubnetGroupResourceType, got[1].ResourceType())
			},
		},
		{
			test:    "cannot list db subnet groups",
			dirName: "aws_rds_db_subnet_group_list",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllDBSubnetGroups").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDbSubnetGroupResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDbSubnetGroupResourceType, resourceaws.AwsDbSubnetGroupResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockRDSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.RDSRepository = fakeRepo
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
				repo = repository.NewRDSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRDSDBSubnetGroupEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestRDSCluster(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockRDSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no cluster",
			dirName: "aws_rds_cluster_empty",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBClusters").Return([]*rds.DBCluster{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "should return results",
			dirName: "aws_rds_clusters_results",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDBClusters").Return([]*rds.DBCluster{
					{
						DBClusterIdentifier: awssdk.String("aurora-cluster-demo"),
						DatabaseName:        awssdk.String("mydb"),
					},
					{
						DBClusterIdentifier: awssdk.String("aurora-cluster-demo-2"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "aurora-cluster-demo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsRDSClusterResourceType, got[0].ResourceType())

				assert.Equal(t, "aurora-cluster-demo-2", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsRDSClusterResourceType, got[1].ResourceType())
			},
		},
		{
			test:    "cannot list clusters",
			dirName: "aws_rds_cluster_denied",
			mocks: func(repository *repository.MockRDSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, "")
				repository.On("ListAllDBClusters").Return(nil, awsError).Once()

				alerter.On("SendAlert", resourceaws.AwsRDSClusterResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRDSClusterResourceType, resourceaws.AwsRDSClusterResourceType), alerts.EnumerationPhase)).Return().Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockRDSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.RDSRepository = fakeRepo
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
				repo = repository.NewRDSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRDSClusterEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
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

func TestRDSDBInstance(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockRDSRepository)
		wantErr error
	}{
		{
			test:    "no db instances",
			dirName: "aws_rds_db_instance_empty",
			mocks: func(repository *repository.MockRDSRepository) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{}, nil)
			},
		},
		{
			test:    "single db instance",
			dirName: "aws_rds_db_instance_single",
			mocks: func(repository *repository.MockRDSRepository) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001")},
				}, nil)
			},
		},
		{
			test:    "multiple mixed db instances",
			dirName: "aws_rds_db_instance_multiple",
			mocks: func(repository *repository.MockRDSRepository) {
				repository.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001")},
					{DBInstanceIdentifier: awssdk.String("database-1")},
				}, nil)
			},
		},
		{
			test:    "cannot list db instances",
			dirName: "aws_rds_db_instance_list",
			mocks: func(repository *repository.MockRDSRepository) {
				repository.On("ListAllDBInstances").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
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
			fakeRepo := &repository.MockRDSRepository{}
			c.mocks(fakeRepo)
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
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDbInstanceResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsDbInstanceResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDbInstanceResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

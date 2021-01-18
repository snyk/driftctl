package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2EbsSnapshotSupplier_Resources(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		snapshotsPages mocks.DescribeSnapshotsPagesOutput
		err            error
	}{
		{
			test:    "no snapshots",
			dirName: "ec2_ebs_snapshot_empty",
			snapshotsPages: mocks.DescribeSnapshotsPagesOutput{
				{
					true,
					&ec2.DescribeSnapshotsOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "with snapshots",
			dirName: "ec2_ebs_snapshot_multiple",
			snapshotsPages: mocks.DescribeSnapshotsPagesOutput{
				{
					false,
					&ec2.DescribeSnapshotsOutput{
						Snapshots: []*ec2.Snapshot{
							{
								SnapshotId: aws.String("snap-0c509a2a880d95a39"),
							},
						},
					},
				},
				{
					true,
					&ec2.DescribeSnapshotsOutput{
						Snapshots: []*ec2.Snapshot{
							{
								SnapshotId: aws.String("snap-00672558cecd93a61"),
							},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewEC2EbsSnapshotSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2EbsSnapshotDeserializer()
			s := &EC2EbsSnapshotSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSEC2EbsSnapshotClient(tt.snapshotsPages),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}

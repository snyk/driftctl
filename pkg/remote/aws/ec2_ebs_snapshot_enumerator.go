package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2EbsSnapshotEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2EbsSnapshotEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2EbsSnapshotEnumerator {
	return &EC2EbsSnapshotEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2EbsSnapshotEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEbsSnapshotResourceType
}

func (e *EC2EbsSnapshotEnumerator) Enumerate() ([]resource.Resource, error) {
	snapshots, err := e.repository.ListAllSnapshots()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(snapshots))

	for _, snapshot := range snapshots {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*snapshot.SnapshotId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

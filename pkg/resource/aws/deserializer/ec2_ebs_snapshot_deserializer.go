package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2EbsSnapshotDeserializer struct {
}

func NewEC2EbsSnapshotDeserializer() *EC2EbsSnapshotDeserializer {
	return &EC2EbsSnapshotDeserializer{}
}

func (s EC2EbsSnapshotDeserializer) HandledType() resource.ResourceType {
	return aws.AwsEbsSnapshotResourceType
}

func (s EC2EbsSnapshotDeserializer) Deserialize(snapshotList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawSnapshot := range snapshotList {
		snapshot, err := decodeEC2EbsSnapshot(rawSnapshot)
		if err != nil {
			logrus.Warnf("error when reading snapshot %s : %+v", snapshot, err)
			return nil, err
		}
		resources = append(resources, snapshot)
	}
	return resources, nil
}

func decodeEC2EbsSnapshot(rawSnapshot cty.Value) (resource.Resource, error) {
	var decodedSnapshot aws.AwsEbsSnapshot
	if err := gocty.FromCtyValue(rawSnapshot, &decodedSnapshot); err != nil {
		return nil, err
	}
	decodedSnapshot.CtyVal = &rawSnapshot
	return &decodedSnapshot, nil
}

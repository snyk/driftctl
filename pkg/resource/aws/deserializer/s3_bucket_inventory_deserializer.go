package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketInventoryDeserializer struct {
}

func NewS3BucketInventoryDeserializer() *S3BucketInventoryDeserializer {
	return &S3BucketInventoryDeserializer{}
}

func (s S3BucketInventoryDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketInventoryResourceType
}

func (s S3BucketInventoryDeserializer) Deserialize(rawInventory []cty.Value) ([]resource.Resource, error) {
	var inventories []resource.Resource
	for _, inventory := range rawInventory {
		inventory := inventory
		var inv aws.AwsS3BucketInventory
		if err := gocty.FromCtyValue(inventory, &inv); err == nil {
			inv.CtyVal = &inventory
			inventories = append(inventories, &inv)
		} else {
			logrus.Warnf("Cannot read s3 bucket inventory %s: %+v", inventory.GoString(), err)
		}
	}
	return inventories, nil
}

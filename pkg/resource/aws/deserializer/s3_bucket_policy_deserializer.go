package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketPolicyDeserializer struct {
}

func NewS3BucketPolicyDeserializer() *S3BucketPolicyDeserializer {
	return &S3BucketPolicyDeserializer{}
}

func (s S3BucketPolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketPolicyResourceType
}

func (s S3BucketPolicyDeserializer) Deserialize(rawPolicy []cty.Value) ([]resource.Resource, error) {
	var inventories []resource.Resource
	for _, policy := range rawPolicy {
		var pol aws.AwsS3BucketPolicy
		if err := gocty.FromCtyValue(policy, &pol); err == nil {
			inventories = append(inventories, &pol)
		} else {
			logrus.Warnf("Cannot read s3 bucket policy %s: %+v", policy.GoString(), err)
		}
	}
	return inventories, nil
}

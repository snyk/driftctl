package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketDeserializer struct {
}

func NewS3BucketDeserializer() *S3BucketDeserializer {
	return &S3BucketDeserializer{}
}

func (s *S3BucketDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketResourceType
}

func (s S3BucketDeserializer) Deserialize(bucketList []cty.Value) ([]resource.Resource, error) {
	buckets := make([]resource.Resource, 0)

	for _, rawBucket := range bucketList {
		bucket, err := decodeS3Bucket(rawBucket)
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

func decodeS3Bucket(rawBucket cty.Value) (resource.Resource, error) {
	var inBucket aws.AwsS3Bucket
	err := gocty.FromCtyValue(rawBucket, &inBucket)
	inBucket.CtyVal = &rawBucket
	return &inBucket, err
}

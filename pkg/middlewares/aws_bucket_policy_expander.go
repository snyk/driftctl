package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Explodes policy found in aws_s3_bucket.policy from state resources to dedicated resources
type AwsBucketPolicyExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsBucketPolicyExpander(resourceFactory resource.ResourceFactory) AwsBucketPolicyExpander {
	return AwsBucketPolicyExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AwsBucketPolicyExpander) Execute(_, resourcesFromState *[]resource.Resource) error {
	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than s3_bucket
		if res.TerraformType() != aws.AwsS3BucketResourceType {
			newList = append(newList, res)
			continue
		}

		bucket, _ := res.(*resource.AbstractResource)
		newList = append(newList, res)

		if hasPolicyAttached(res.TerraformId(), resourcesFromState) {
			bucket.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.handlePolicy(bucket, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsBucketPolicyExpander) handlePolicy(bucket *resource.AbstractResource, results *[]resource.Resource) error {
	policyAttr, exist := bucket.Attrs.Get("policy")
	if !exist || policyAttr == nil || policyAttr == "" {
		return nil
	}
	policy := policyAttr.(string)

	bucketAttr, _ := bucket.Attrs.Get("bucket")
	bucketName := bucketAttr.(string)
	data := map[string]interface{}{
		"id":     bucket.TerraformId(),
		"bucket": (*bucket.Attrs)["bucket"],
		"policy": (*bucket.Attrs)["policy"],
	}
	ctyVal, err := m.resourceFactory.CreateResource(data, "aws_s3_bucket_policy")
	if err != nil {
		return err
	}

	newPolicy := &aws.AwsS3BucketPolicy{
		Id:     bucket.TerraformId(),
		Bucket: &bucketName,
		Policy: &policy,
		CtyVal: ctyVal,
	}
	normalizedRes, err := newPolicy.NormalizeForState()
	if err != nil {
		return err
	}
	*results = append(*results, normalizedRes)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from bucket")

	bucket.Attrs.SafeDelete([]string{"policy"})
	return nil
}

// Return true if the bucket has a aws_bucket_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_bucket with an inline policy
// AND a aws_bucket_policy resource at the same time. At the end, on the AWS console,
// the aws_bucket_policy will be used.
func hasPolicyAttached(bucket string, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsS3BucketPolicyResourceType &&
			res.TerraformId() == bucket {
			return true
		}
	}
	return false
}

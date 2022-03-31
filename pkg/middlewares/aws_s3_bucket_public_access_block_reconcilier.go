package middlewares

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsS3BucketPublicAccessBlockReconciler middleware ignores every s3 bucket public block that is set to the default values (every option set to false)
// This is used to avoid displaying false positive unmanaged resources.
// The problem here is that the aws SDK can either return an error `NoSuchPublicAccessBlockConfiguration` while
// retrieving bucket public block, or a response with all fields set to false (the default)
//
// To reproduce this edgy case you can do that:
// - Disable this middleware
// - Go to the folder of the test `TestAcc_Aws_S3Bucket_PublicAccessBlock` : `testdata/acc/aws_s3_bucket_public_access_block`
// - Apply tf code
// - Run a scan with the driftignore from the test folder (ignore everything but bucket and public access block)
//   - Infra should be in sync (be sure that you have no dangling bucket in your aws test env)
// - Create a new unmanaged bucket from the console, with every option from the policy block set to false
// - Run the scan again
//   - One resource should be unmanaged: the bucket (expected behavior)
// - Go to the console and update public access block for that bucket
// - Run the scan again
//   - We should now have a new public access block resource unmanaged (expected)
// - Now uncheck back all things in the public block you just updated
// - Run the scan again
//   - We still have the public block as unmanaged, this is NOT expected since all values are back to default
//
// This simple middleware is handling that edge case by removing resource that have every attribute set to false both
// from resources from state and resources from remote.
type AwsS3BucketPublicAccessBlockReconciler struct{}

func NewAwsS3BucketPublicAccessBlockReconciler() *AwsS3BucketPublicAccessBlockReconciler {
	return &AwsS3BucketPublicAccessBlockReconciler{}
}

func (r AwsS3BucketPublicAccessBlockReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)
	for _, res := range *remoteResources {
		// Only keep non-default public access blocks
		if r.isDefaultPublicAccessBlock(res) {
			logrus.WithField("id", res.ResourceId()).Debug("Ignored default aws_s3_bucket_public_access_block from remote")
			continue
		}
		newRemoteResources = append(newRemoteResources, res)
	}
	*remoteResources = newRemoteResources

	newResourcesFromState := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Only keep non-default public access blocks
		if r.isDefaultPublicAccessBlock(res) {
			logrus.WithField("id", res.ResourceId()).Debug("Ignored default aws_s3_bucket_public_access_block from state")
			continue
		}
		newResourcesFromState = append(newResourcesFromState, res)
	}
	*resourcesFromState = newResourcesFromState

	return nil
}

func (r AwsS3BucketPublicAccessBlockReconciler) isDefaultPublicAccessBlock(res *resource.Resource) bool {

	if res.ResourceType() != aws.AwsS3BucketPublicAccessBlockResourceType {
		return false
	}

	if !awssdk.BoolValue(res.Attributes().GetBool("block_public_acls")) &&
		!awssdk.BoolValue(res.Attributes().GetBool("block_public_policy")) &&
		!awssdk.BoolValue(res.Attributes().GetBool("ignore_public_acls")) &&
		!awssdk.BoolValue(res.Attributes().GetBool("restrict_public_buckets")) {
		return true
	}

	return false
}

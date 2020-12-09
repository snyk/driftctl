package middlewares

import (
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Remove grant field on remote resources when acl field != private in state
type S3BucketAcl struct{}

func NewS3BucketAcl() S3BucketAcl {
	return S3BucketAcl{}
}

func (m S3BucketAcl) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	for _, iacResource := range *resourcesFromState {
		// Ignore all resources other than s3 buckets
		if iacResource.TerraformType() != aws.AwsS3BucketResourceType {
			continue
		}

		decodedIacResource, _ := iacResource.(*aws.AwsS3Bucket)

		for _, remoteResource := range *remoteResources {
			if resource.IsSameResource(remoteResource, decodedIacResource) {
				decodedRemoteResource, _ := remoteResource.(*aws.AwsS3Bucket)
				if decodedIacResource.Acl != nil && *decodedIacResource.Acl != "private" {
					logrus.WithFields(logrus.Fields{
						"type": decodedRemoteResource.TerraformType(),
						"id":   decodedRemoteResource.TerraformId(),
					}).Debug("Found a resource to update")
					// Use reflection to reset to zero value
					reflect.ValueOf(decodedRemoteResource.Grant).Elem().Set(
						reflect.Zero(
							reflect.ValueOf(*decodedRemoteResource.Grant).Type(),
						),
					)
				}
				break
			}
		}
	}

	return nil
}

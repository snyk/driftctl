package middlewares

import (
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

		decodedIacResource, _ := iacResource.(*resource.AbstractResource)

		for _, remoteResource := range *remoteResources {
			if resource.IsSameResource(remoteResource, decodedIacResource) {
				decodedRemoteResource, _ := remoteResource.(*resource.AbstractResource)
				aclAttr, exist := decodedIacResource.Attrs.Get("acl")
				if !exist || aclAttr == nil || aclAttr == "" {
					break
				}
				if aclAttr != "private" {
					logrus.WithFields(logrus.Fields{
						"type": decodedRemoteResource.TerraformType(),
						"id":   decodedRemoteResource.TerraformId(),
					}).Debug("Found a resource to update")
					decodedRemoteResource.Attrs.SafeDelete([]string{"grant"})
				}
				break
			}
		}

		decodedIacResource.Attrs.SafeDelete([]string{"acl"})
	}

	return nil
}

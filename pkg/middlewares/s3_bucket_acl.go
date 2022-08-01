package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Remove grant field on remote resources when acl field != private in state
type S3BucketAcl struct{}

func NewS3BucketAcl() S3BucketAcl {
	return S3BucketAcl{}
}

func (m S3BucketAcl) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	for _, iacResource := range *resourcesFromState {
		// Ignore all resources other than s3 buckets
		if iacResource.ResourceType() != aws.AwsS3BucketResourceType {
			continue
		}

		for _, remoteResource := range *remoteResources {
			if remoteResource.Equal(iacResource) {
				aclAttr, exist := iacResource.Attrs.Get("acl")
				if !exist || aclAttr == nil || aclAttr == "" {
					break
				}
				if aclAttr != "private" {
					logrus.WithFields(logrus.Fields{
						"type": remoteResource.ResourceType(),
						"id":   remoteResource.ResourceId(),
					}).Debug("Found a resource to update")
					remoteResource.Attrs.SafeDelete([]string{"grant"})
				}
				break
			}
		}

		iacResource.Attrs.SafeDelete([]string{"acl"})
	}

	return nil
}

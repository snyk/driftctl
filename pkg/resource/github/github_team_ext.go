package github

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *GithubTeam) Attributes() map[string]string {
	attrs := make(map[string]string)
	attrs["Id"] = r.Id
	if r.Name != nil && *r.Name != "" {
		attrs["Name"] = *r.Name
	}
	return attrs
}

func (r *GithubTeam) NormalizeForState() (resource.Resource, error) {
	if r.CreateDefaultMaintainer == nil {
		r.CreateDefaultMaintainer = awssdk.Bool(false)
	}
	return r, nil
}

func (r *GithubTeam) NormalizeForProvider() (resource.Resource, error) {
	if r.CreateDefaultMaintainer == nil {
		r.CreateDefaultMaintainer = awssdk.Bool(false)
	}
	return r, nil
}

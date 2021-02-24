package github

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *GithubTeam) String() string {
	if r.Name != nil {
		return fmt.Sprintf("%s (Id: %s)", *r.Name, r.Id)
	}
	return r.Id
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

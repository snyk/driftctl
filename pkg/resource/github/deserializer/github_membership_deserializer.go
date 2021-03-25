package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type GithubMembershipDeserializer struct {
}

func NewGithubMembershipDeserializer() *GithubMembershipDeserializer {
	return &GithubMembershipDeserializer{}
}

func (s GithubMembershipDeserializer) HandledType() resource.ResourceType {
	return github.GithubMembershipResourceType
}

func (s GithubMembershipDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeMembership(&raw)
		if err != nil {
			return nil, errors.Wrapf(err, "error when deserializing github_membership %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeMembership(res *cty.Value) (resource.Resource, error) {
	var decoded github.GithubMembership
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = res
	return &decoded, nil
}

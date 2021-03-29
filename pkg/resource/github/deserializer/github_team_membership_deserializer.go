package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type GithubTeamMembershipDeserializer struct {
}

func NewGithubTeamMembershipDeserializer() *GithubTeamMembershipDeserializer {
	return &GithubTeamMembershipDeserializer{}
}

func (s GithubTeamMembershipDeserializer) HandledType() resource.ResourceType {
	return github.GithubTeamMembershipResourceType
}

func (s GithubTeamMembershipDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeTeamMembership(&raw)
		if err != nil {
			return nil, errors.Wrapf(err, "error when deserializing github_team_membership %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeTeamMembership(res *cty.Value) (resource.Resource, error) {
	var decoded github.GithubTeamMembership
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = res
	return &decoded, nil
}

package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type GithubTeamDeserializer struct {
}

func NewGithubTeamDeserializer() *GithubTeamDeserializer {
	return &GithubTeamDeserializer{}
}

func (s GithubTeamDeserializer) HandledType() resource.ResourceType {
	return github.GithubTeamResourceType
}

func (s GithubTeamDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeTeam(&raw)
		if err != nil {
			return nil, errors.Wrapf(err, "error when deserializing github_team %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeTeam(res *cty.Value) (resource.Resource, error) {
	var decoded github.GithubTeam
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = res
	return &decoded, nil
}

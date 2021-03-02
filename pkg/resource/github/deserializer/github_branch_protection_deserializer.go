package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type GithubBranchProtectionDeserializer struct {
}

func NewGithubBranchProtectionDeserializer() *GithubBranchProtectionDeserializer {
	return &GithubBranchProtectionDeserializer{}
}

func (s GithubBranchProtectionDeserializer) HandledType() resource.ResourceType {
	return github.GithubBranchProtectionResourceType
}

func (s GithubBranchProtectionDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeBranchProtection(&raw)
		if err != nil {
			return nil, errors.Wrapf(err, "error when deserializing github_branch_protection %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeBranchProtection(res *cty.Value) (resource.Resource, error) {
	var decoded github.GithubBranchProtection
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

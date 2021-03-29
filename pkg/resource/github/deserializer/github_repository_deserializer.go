package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type GithubRepositoryDeserializer struct {
}

func NewGithubRepositoryDeserializer() *GithubRepositoryDeserializer {
	return &GithubRepositoryDeserializer{}
}

func (s GithubRepositoryDeserializer) HandledType() resource.ResourceType {
	return github.GithubRepositoryResourceType
}

func (s GithubRepositoryDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeRepository(&raw)
		if err != nil {
			return nil, errors.Wrapf(err, "error when deserializing github_repository %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeRepository(res *cty.Value) (resource.Resource, error) {
	var decoded github.GithubRepository
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = res
	return &decoded, nil
}

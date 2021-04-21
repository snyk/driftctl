package resource

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/dctlcty"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
)

func fetchNestedBlocks(root string, metadata map[string]dctlcty.AttributeMetadata, block map[string]*configschema.NestedBlock) {
	for s, nestedBlock := range block {
		path := s
		if root != "" {
			path = strings.Join([]string{root, s}, ".")
		}
		for s2, attr := range nestedBlock.Attributes {
			nestedPath := strings.Join([]string{path, s2}, ".")
			metadata[nestedPath] = dctlcty.AttributeMetadata{
				Configshema: *attr,
			}
		}
		fetchNestedBlocks(path, metadata, nestedBlock.BlockTypes)
	}
}

func RetrieveAttributesFromSchemas(schema map[string]providers.Schema) {
	for typ, sch := range schema {
		attributeMetas := map[string]dctlcty.AttributeMetadata{}
		for s, attribute := range sch.Block.Attributes {
			attributeMetas[s] = dctlcty.AttributeMetadata{
				Configshema: *attribute,
			}
		}

		fetchNestedBlocks("", attributeMetas, sch.Block.BlockTypes)

		dctlcty.AddMetadata(typ, &dctlcty.ResourceMetadata{
			AttributeMetadata: attributeMetas,
		})
	}
}

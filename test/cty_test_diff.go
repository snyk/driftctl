package test

import (
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/zclconf/go-cty/cty/json"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func CtyTestDiff(got []resource.Resource, dirName string, provider terraform.TerraformProvider, deserializer deserializer.CTYDeserializer, shouldUpdate bool, t *testing.T) {
	resGoldenName := "results.golden.json"
	ctyType := cty.List(provider.Schema()[deserializer.HandledType().String()].Block.ImpliedType())
	if shouldUpdate {
		ctVal, err := gocty.ToCtyValue(got, ctyType)
		if err != nil {
			panic(err)
		}
		unm, err := json.Marshal(ctVal, ctyType)
		if err != nil {
			panic(err)
		}
		goldenfile.WriteFile(dirName, unm, resGoldenName)
	}
	file := goldenfile.ReadFile(dirName, resGoldenName)
	decodedJson, err := json.Unmarshal(file, ctyType)
	if err != nil {
		panic(err)
	}
	expectedResources, err := deserializer.Deserialize(decodedJson.AsValueSlice())
	if err != nil {
		panic(err)
	}

	changelog, err := diff.Diff(got, expectedResources)
	if err != nil {
		panic(err)
	}
	if len(changelog) > 0 {
		for _, change := range changelog {
			t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
		}
	}
}

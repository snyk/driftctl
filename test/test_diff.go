package test

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/r3labs/diff/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

func TestAgainstGoldenFile(
	got []resource.Resource,
	ty string,
	dirName string,
	provider terraform.TerraformProvider,
	deserializer *resource.Deserializer,
	shouldUpdate bool,
	tt *testing.T,
) {
	resGoldenName := "results.golden.json"
	var expectedResources []resource.Resource
	ctyType := cty.List(provider.Schema()[ty].Block.ImpliedType())

	// update golden file
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

	// read golden file
	file := goldenfile.ReadFile(dirName, resGoldenName)
	decodedJson, err := json.Unmarshal(file, ctyType)
	if err != nil {
		panic(err)
	}
	decodedResources, err := deserializer.Deserialize(ty, decodedJson.AsValueSlice())
	if err != nil {
		panic(err)
	}
	expectedResources = append(expectedResources, decodedResources...)

	// diff
	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		panic(err)
	}

	got = resource.Sort(got)
	expectedResources = resource.Sort(expectedResources)

	changelog, err := differ.Diff(got, expectedResources)

	if err != nil {
		panic(err)
	}
	if len(changelog) > 0 {
		for _, change := range changelog {
			tt.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
		}
	}
}

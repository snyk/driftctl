package test

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

func TestAgainstGoldenFile(
	got []*resource.Resource,
	ty string,
	dirName string,
	provider terraform.TerraformProvider,
	deserializer *resource.Deserializer,
	shouldUpdate bool,
	tt *testing.T,
) {
	var expectedResources []*resource.Resource
	ctyType := cty.List(provider.Schema()[ty].Block.ImpliedType())

	// update golden file
	if shouldUpdate {
		attributes := make([]*resource.Attributes, 0, len(got))
		for _, res := range got {
			attributes = append(attributes, res.Attributes())
		}
		ctVal, err := gocty.ToCtyValue(attributes, ctyType)
		if err != nil {
			panic(err)
		}
		unm, err := json.Marshal(ctVal, ctyType)
		if err != nil {
			panic(err)
		}
		goldenfile.WriteFile(dirName, unm, goldenfile.ResultsFilename)
	}

	// read golden file
	file := goldenfile.ReadFile(dirName, goldenfile.ResultsFilename)
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

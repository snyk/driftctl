package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/zclconf/go-cty/cty/json"

	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func doTestDiff(got []resource.Resource, dirName string, provider terraform.TerraformProvider, deserializer *resource.Deserializer, shouldUpdate bool) (diff.Changelog, error) {
	resources := make(map[string][]resource.Attributes)

	for _, r := range got {
		res, exist := resources[r.TerraformType()]

		if !exist {
			resources[r.TerraformType()] = []resource.Attributes{*r.Attributes()}
			continue
		}

		resources[r.TerraformType()] = append(res, *r.Attributes())
	}

	expectedResources := []resource.Resource{}
	for ty, resList := range resources {
		resGoldenName := "results.golden.json"
		if len(resources) > 1 {
			resGoldenName = fmt.Sprintf("results.%s.golden.json", ty)
		}
		ctyType := cty.List(provider.Schema()[ty].Block.ImpliedType())
		if shouldUpdate {
			ctVal, err := gocty.ToCtyValue(resList, ctyType)
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
		decodedResources, err := deserializer.Deserialize(resource.ResourceType(ty), decodedJson.AsValueSlice())
		if err != nil {
			panic(err)
		}
		expectedResources = append(expectedResources, decodedResources...)
	}

	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		panic(err)
	}

	got = resource.Sort(got)
	expectedResources = resource.Sort(expectedResources)

	return differ.Diff(got, expectedResources)
}

func CtyTestDiff(got []resource.Resource, dirName string, provider terraform.TerraformProvider, deserializer *resource.Deserializer, shouldUpdate bool, t *testing.T) {
	changelog, err := doTestDiff(got, dirName, provider, deserializer, shouldUpdate)
	if err != nil {
		panic(err)
	}
	if len(changelog) > 0 {
		for _, change := range changelog {
			t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
		}
	}
}

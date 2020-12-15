package test

import (
	"fmt"
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

func doTestDiff(got []resource.Resource, dirName string, provider terraform.TerraformProvider, deserializers []deserializer.CTYDeserializer, shouldUpdate bool) (diff.Changelog, error) {

	expectedResources := []resource.Resource{}
	for _, deserializer := range deserializers {
		ty := deserializer.HandledType().String()

		resList := []resource.Resource{}
		for _, res := range got {
			if res.TerraformType() == ty {
				resList = append(resList, res)
			}
		}
		resGoldenName := "results.golden.json"
		if len(deserializers) > 1 {
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
		decodedResources, err := deserializer.Deserialize(decodedJson.AsValueSlice())
		if err != nil {
			panic(err)
		}
		expectedResources = append(expectedResources, decodedResources...)
	}

	return diff.Diff(got, expectedResources)
}

func CtyTestDiff(got []resource.Resource, dirName string, provider terraform.TerraformProvider, d deserializer.CTYDeserializer, shouldUpdate bool, t *testing.T) {
	deserializers := []deserializer.CTYDeserializer{d}
	CtyTestDiffMixed(got, dirName, provider, deserializers, shouldUpdate, t)
}

func CtyTestDiffMixed(got []resource.Resource, dirName string, provider terraform.TerraformProvider, deserializers []deserializer.CTYDeserializer, shouldUpdate bool, t *testing.T) {
	changelog, err := doTestDiff(got, dirName, provider, deserializers, shouldUpdate)
	if err != nil {
		panic(err)
	}
	if len(changelog) > 0 {
		for _, change := range changelog {
			t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
		}
	}
}

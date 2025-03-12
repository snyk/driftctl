package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// That method is used to compare the result of the enumeration with the golden file.
// That method does not use cty and types from the terraform provider to deserialize resources.
// Some resources returned by the enumeration may have missing fields, and if we use cty deserialization we're
// gonna recreate those missing fields to respect the schema.
func TestAgainstGoldenFileNoCty(
	got []*resource.Resource,
	ty string,
	dirName string,
	_ terraform.TerraformProvider,
	_ *resource.Deserializer,
	shouldUpdate bool,
	tt *testing.T) {
	var expectedResources []*resource.Resource

	// update golden file
	if shouldUpdate {
		attributes := make([]*resource.Attributes, 0, len(got))
		for _, res := range got {
			attributes = append(attributes, res.Attributes())
		}
		fileContent, err := json.MarshalIndent(attributes, "", " ")
		if err != nil {
			panic(err)
		}
		goldenfile.WriteFile(dirName, fileContent, goldenfile.ResultsFilename)
	}

	// read golden file
	file := goldenfile.ReadFile(dirName, goldenfile.ResultsFilename)
	rawResources := make([]map[string]interface{}, 0)
	err := json.Unmarshal(file, &rawResources)
	if err != nil {
		panic(err)
	}

	for _, r := range rawResources {
		res := r
		expectedResources = append(expectedResources, &resource.Resource{
			Id:    res["id"].(string),
			Type:  ty,
			Attrs: (*resource.Attributes)(&res),
		})
	}

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

func testAgainstGoldenFileCty(
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
		unm, err := ctyjson.Marshal(ctVal, ctyType)
		if err != nil {
			panic(err)
		}
		goldenfile.WriteFile(dirName, unm, goldenfile.ResultsFilename)
	}

	// read golden file
	file := goldenfile.ReadFile(dirName, goldenfile.ResultsFilename)
	decodedJson, err := ctyjson.Unmarshal(file, ctyType)
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

func TestAgainstGoldenFile(
	got []*resource.Resource,
	ty string,
	dirName string,
	provider terraform.TerraformProvider,
	deserializer *resource.Deserializer,
	shouldUpdate bool,
	tt *testing.T,
) {
	testAgainstGoldenFileCty(got, ty, dirName, provider, deserializer, shouldUpdate, tt)
}

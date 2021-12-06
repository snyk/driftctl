package terraform

import (
	"os"
	"testing"

	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test/schemas"
)

// You can use this test function to create a schema file for a given provider in a given version
// You may want to update part of this code to change provider and version to generate desired schema
// To use this test you should run this command from the repository root
// DCTL_UPDATE_TEST_SCHEMA=true go test ./test/terraform
// You may need to setup proper environment variable to make the terraform provider work
// DCTL_UPDATE_TEST_SCHEMA=true AWS_PROFILE=myprofile go test ./test/terraform
func TestCreateNewSchema(t *testing.T) {

	if os.Getenv("DCTL_UPDATE_TEST_SCHEMA") != "true" {
		t.SkipNow()
	}

	providerLibrary := terraform.NewProviderLibrary()

	// Replace this with provider you want to create schema
	realProvider, _ := InitTestAwsProvider(providerLibrary, "3.19.0")

	err := realProvider.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = schemas.WriteTestSchema(realProvider.Schema(), realProvider.Name(), realProvider.Version())
	if err != nil {
		t.Fatal(err)
	}
}

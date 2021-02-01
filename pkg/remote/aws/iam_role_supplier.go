package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

var iamRoleExclusionList = map[string]struct{}{
	// Enabled by default for aws to enable support, not removable
	"AWSServiceRoleForSupport": {},
	// Enabled and not removable for every org account
	"AWSServiceRoleForOrganizations": {},
	// Not manageable by IaC and set by default
	"AWSServiceRoleForTrustedAdvisor": {},
}

type IamRoleSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamRoleSupplier(provider *TerraformProvider) *IamRoleSupplier {
	return &IamRoleSupplier{
		provider,
		awsdeserializer.NewIamRoleDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func awsIamRoleShouldBeIgnored(roleName string) bool {
	_, ok := iamRoleExclusionList[roleName]
	return ok
}

func (s IamRoleSupplier) Resources() ([]resource.Resource, error) {
	roles, err := listIamRoles(s.client, resourceaws.AwsIamRoleResourceType)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(roles) > 0 {
		for _, role := range roles {
			u := *role
			if u.RoleName != nil && awsIamRoleShouldBeIgnored(*u.RoleName) {
				continue
			}
			s.runner.Run(func() (cty.Value, error) {
				return s.readRes(&u)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s IamRoleSupplier) readRes(resource *iam.Role) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamRoleResourceType,
			ID: *resource.RoleName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam role %s[%s]: %+v", *resource.RoleName, resourceaws.AwsIamRoleResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamRoles(client iamiface.IAMAPI, supplierType string) ([]*iam.Role, error) {
	var resources []*iam.Role
	input := &iam.ListRolesInput{}
	err := client.ListRolesPages(input, func(res *iam.ListRolesOutput, lastPage bool) bool {
		resources = append(resources, res.Roles...)
		return !lastPage
	})
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, supplierType, resourceaws.AwsIamRoleResourceType)
	}
	return resources, nil
}

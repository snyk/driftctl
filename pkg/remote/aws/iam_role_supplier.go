package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

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
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamRoleSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamRoleSupplier {
	return &IamRoleSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func awsIamRoleShouldBeIgnored(roleName string) bool {
	_, ok := iamRoleExclusionList[roleName]
	return ok
}

func (s *IamRoleSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamRoleResourceType
}

func (s *IamRoleSupplier) Resources() ([]resource.Resource, error) {
	roles, err := s.repo.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(roles) > 0 {
		for _, role := range roles {
			u := *role
			if u.RoleName != nil && awsIamRoleShouldBeIgnored(*u.RoleName) {
				continue
			}
			s.runner.Run(func() (cty.Value, error) {
				return s.readRole(&u)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamRoleSupplier) readRole(resource *iam.Role) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: *resource.RoleName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam role %s[%s]: %+v", *resource.RoleName, s.SuppliedType(), err)
		return cty.NilVal, err
	}

	return *res, nil
}

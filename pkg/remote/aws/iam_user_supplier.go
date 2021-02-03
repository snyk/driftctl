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

type IamUserSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamUserSupplier(provider *TerraformProvider) *IamUserSupplier {
	return &IamUserSupplier{
		provider,
		awsdeserializer.NewIamUserDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamUserSupplier) Resources() ([]resource.Resource, error) {
	users, err := listIamUsers(s.client, resourceaws.AwsIamUserResourceType)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(users) > 0 {
		for _, user := range users {
			u := *user
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

func (s IamUserSupplier) readRes(user *iam.User) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserResourceType,
			ID: *user.UserName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user %s[%s]: %+v", *user.UserName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamUsers(client iamiface.IAMAPI, supplierType string) ([]*iam.User, error) {
	var resources []*iam.User
	input := &iam.ListUsersInput{}
	err := client.ListUsersPages(input, func(res *iam.ListUsersOutput, lastPage bool) bool {
		resources = append(resources, res.Users...)
		return !lastPage
	})
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, supplierType, resourceaws.AwsIamUserResourceType)
	}
	return resources, nil
}

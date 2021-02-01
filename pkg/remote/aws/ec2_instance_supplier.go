package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2InstanceSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2InstanceSupplier(provider *TerraformProvider) *EC2InstanceSupplier {
	return &EC2InstanceSupplier{
		provider,
		awsdeserializer.NewEC2InstanceDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s EC2InstanceSupplier) Resources() ([]resource.Resource, error) {
	instances, err := listInstances(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsInstanceResourceType)
	}

	results := make([]cty.Value, 0)
	if len(instances) > 0 {
		for _, instance := range instances {
			id := aws.StringValue(instance.InstanceId)
			s.runner.Run(func() (cty.Value, error) {
				return s.readInstance(id)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2InstanceSupplier) readInstance(id string) (cty.Value, error) {
	resInstance, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsInstanceResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading instance %s: %+v", id, err)
		return cty.NilVal, err
	}
	if resInstance.IsNull() {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("Instance read returned nil (instance may be terminated), ignoring ...")
	}
	return *resInstance, nil
}

func listInstances(client ec2iface.EC2API) ([]*ec2.Instance, error) {
	var instances []*ec2.Instance
	input := &ec2.DescribeInstancesInput{}
	err := client.DescribeInstancesPages(input, func(res *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range res.Reservations {
			instances = append(instances, reservation.Instances...)
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return instances, nil
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg"
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

type EC2KeyPairSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2KeyPairSupplier(runner *pkg.ParallelRunner, client ec2iface.EC2API) *EC2KeyPairSupplier {
	return &EC2KeyPairSupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewEC2KeyPairDeserializer(), client, terraform.NewParallelResourceReader(runner)}
}

func (s EC2KeyPairSupplier) Resources() ([]resource.Resource, error) {
	input := &ec2.DescribeKeyPairsInput{}
	response, err := s.client.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(response.KeyPairs) > 0 {
		for _, kp := range response.KeyPairs {
			name := aws.StringValue(kp.KeyName)
			s.runner.Run(func() (cty.Value, error) {
				return s.readKeyPair(name)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2KeyPairSupplier) readKeyPair(name string) (cty.Value, error) {
	resKp, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsKeyPairResourceType,
		ID: name,
	})
	if err != nil {
		logrus.Warnf("Error reading key pair %s: %+v", name, err)
		return cty.NilVal, err
	}
	return *resKp, nil
}

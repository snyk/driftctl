package mocks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type DescribeInstancesPagesOutput []struct {
	LastPage bool
	Response *ec2.DescribeInstancesOutput
}

type DescribeSnapshotsPagesOutput []struct {
	LastPage bool
	Response *ec2.DescribeSnapshotsOutput
}

type DescribeVolumesPagesOutput []struct {
	LastPage bool
	Response *ec2.DescribeVolumesOutput
}

type DescribeSecurityGroupsPagesOutput []struct {
	LastPage bool
	Response *ec2.DescribeSecurityGroupsOutput
}

type MockAWSEC2Client struct {
	ec2iface.EC2API
	instancesPages      DescribeInstancesPagesOutput
	kpNames             []string
	amiIDs              []string
	volumesPages        DescribeVolumesPagesOutput
	snapshotsPages      DescribeSnapshotsPagesOutput
	addresses           []*ec2.Address
	securityGroupsPages DescribeSecurityGroupsPagesOutput
	err                 error
}

func NewMockAWSEC2ErrorClient(err error) *MockAWSEC2Client {
	return &MockAWSEC2Client{err: err}
}

func NewMockAWSEC2InstanceClient(instancesPages DescribeInstancesPagesOutput) *MockAWSEC2Client {
	return &MockAWSEC2Client{instancesPages: instancesPages}
}

func NewMockAWSEC2KeyPairClient(kpNames []string) *MockAWSEC2Client {
	return &MockAWSEC2Client{kpNames: kpNames}
}

func NewMockAWSEC2AmiClient(amiIDs []string) *MockAWSEC2Client {
	return &MockAWSEC2Client{amiIDs: amiIDs}
}

func NewMockAWSEC2EbsVolumeClient(volumesPages DescribeVolumesPagesOutput) *MockAWSEC2Client {
	return &MockAWSEC2Client{volumesPages: volumesPages}
}

func NewMockAWSEC2EbsSnapshotClient(snapshotsPages DescribeSnapshotsPagesOutput) *MockAWSEC2Client {
	return &MockAWSEC2Client{snapshotsPages: snapshotsPages}
}

func NewMockAWSEC2EipClient(addresses []*ec2.Address) *MockAWSEC2Client {
	return &MockAWSEC2Client{addresses: addresses}
}

func NewMockAWSVPCSecurityGroupClient(securityGroupsPages DescribeSecurityGroupsPagesOutput) *MockAWSEC2Client {
	return &MockAWSEC2Client{securityGroupsPages: securityGroupsPages}
}

func (m *MockAWSEC2Client) DescribeInstancesPages(_ *ec2.DescribeInstancesInput, cb func(*ec2.DescribeInstancesOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, instancesPage := range m.instancesPages {
		cb(instancesPage.Response, instancesPage.LastPage)
	}
	return nil
}

func (m *MockAWSEC2Client) DescribeKeyPairs(*ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	var kps []*ec2.KeyPairInfo
	for _, name := range m.kpNames {
		kps = append(kps, &ec2.KeyPairInfo{
			KeyName: aws.String(name),
		})
	}
	return &ec2.DescribeKeyPairsOutput{
		KeyPairs: kps,
	}, nil
}

func (m *MockAWSEC2Client) DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	var images []*ec2.Image
	for _, id := range m.amiIDs {
		images = append(images, &ec2.Image{
			ImageId: aws.String(id),
		})
	}
	return &ec2.DescribeImagesOutput{
		Images: images,
	}, nil
}

func (m *MockAWSEC2Client) DescribeVolumesPages(_ *ec2.DescribeVolumesInput, cb func(*ec2.DescribeVolumesOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, volumesPage := range m.volumesPages {
		cb(volumesPage.Response, volumesPage.LastPage)
	}
	return nil
}

func (m *MockAWSEC2Client) DescribeSnapshotsPages(_ *ec2.DescribeSnapshotsInput, cb func(*ec2.DescribeSnapshotsOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, snapshotsPage := range m.snapshotsPages {
		cb(snapshotsPage.Response, snapshotsPage.LastPage)
	}
	return nil
}

func (m *MockAWSEC2Client) DescribeAddresses(*ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	addresses := m.addresses
	return &ec2.DescribeAddressesOutput{
		Addresses: addresses,
	}, nil
}

func (m *MockAWSEC2Client) DescribeSecurityGroupsPages(_ *ec2.DescribeSecurityGroupsInput, cb func(*ec2.DescribeSecurityGroupsOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, securityGroupsPage := range m.securityGroupsPages {
		cb(securityGroupsPage.Response, securityGroupsPage.LastPage)
	}
	return nil
}

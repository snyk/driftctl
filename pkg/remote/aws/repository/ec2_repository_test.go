package repository

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ec2Repository_ListAllImages(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.Image
		wantErr error
	}{
		{
			name: "List all images",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeImages",
					&ec2.DescribeImagesInput{
						Owners: []*string{
							aws.String("self"),
						},
					}).Return(&ec2.DescribeImagesOutput{
					Images: []*ec2.Image{
						{ImageId: aws.String("1")},
						{ImageId: aws.String("2")},
						{ImageId: aws.String("3")},
						{ImageId: aws.String("4")},
					},
				}, nil).Once()
			},
			want: []*ec2.Image{
				{ImageId: aws.String("1")},
				{ImageId: aws.String("2")},
				{ImageId: aws.String("3")},
				{ImageId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllImages()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllImages()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllSnapshots(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.Snapshot
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeSnapshotsPages",
					&ec2.DescribeSnapshotsInput{
						OwnerIds: []*string{
							aws.String("self"),
						},
					},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSnapshotsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSnapshotsOutput{
							Snapshots: []*ec2.Snapshot{
								{VolumeId: aws.String("1")},
								{VolumeId: aws.String("2")},
								{VolumeId: aws.String("3")},
								{VolumeId: aws.String("4")},
							},
						}, false)
						callback(&ec2.DescribeSnapshotsOutput{
							Snapshots: []*ec2.Snapshot{
								{VolumeId: aws.String("5")},
								{VolumeId: aws.String("6")},
								{VolumeId: aws.String("7")},
								{VolumeId: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Snapshot{
				{VolumeId: aws.String("1")},
				{VolumeId: aws.String("2")},
				{VolumeId: aws.String("3")},
				{VolumeId: aws.String("4")},
				{VolumeId: aws.String("5")},
				{VolumeId: aws.String("6")},
				{VolumeId: aws.String("7")},
				{VolumeId: aws.String("8")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllSnapshots()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllSnapshots()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllVolumes(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.Volume
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeVolumesPages",
					&ec2.DescribeVolumesInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVolumesOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeVolumesOutput{
							Volumes: []*ec2.Volume{
								{VolumeId: aws.String("1")},
								{VolumeId: aws.String("2")},
								{VolumeId: aws.String("3")},
								{VolumeId: aws.String("4")},
							},
						}, false)
						callback(&ec2.DescribeVolumesOutput{
							Volumes: []*ec2.Volume{
								{VolumeId: aws.String("5")},
								{VolumeId: aws.String("6")},
								{VolumeId: aws.String("7")},
								{VolumeId: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Volume{
				{VolumeId: aws.String("1")},
				{VolumeId: aws.String("2")},
				{VolumeId: aws.String("3")},
				{VolumeId: aws.String("4")},
				{VolumeId: aws.String("5")},
				{VolumeId: aws.String("6")},
				{VolumeId: aws.String("7")},
				{VolumeId: aws.String("8")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllVolumes()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllVolumes()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllAddresses(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.Address
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeAddresses", &ec2.DescribeAddressesInput{}).
					Return(&ec2.DescribeAddressesOutput{
						Addresses: []*ec2.Address{
							{AssociationId: aws.String("1")},
							{AssociationId: aws.String("2")},
							{AssociationId: aws.String("3")},
							{AssociationId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []*ec2.Address{
				{AssociationId: aws.String("1")},
				{AssociationId: aws.String("2")},
				{AssociationId: aws.String("3")},
				{AssociationId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllAddresses()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllAddresses()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllAddressesAssociation(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []string
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeAddresses", &ec2.DescribeAddressesInput{}).
					Return(&ec2.DescribeAddressesOutput{
						Addresses: []*ec2.Address{
							{AssociationId: aws.String("1")},
							{AssociationId: aws.String("2")},
							{AssociationId: aws.String("3")},
							{AssociationId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []string{
				"1",
				"2",
				"3",
				"4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllAddressesAssociation()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllAddressesAssociation()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllInstances(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.Instance
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeInstancesPages",
					&ec2.DescribeInstancesInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInstancesOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeInstancesOutput{
							Reservations: []*ec2.Reservation{
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("1")},
										{ImageId: aws.String("2")},
										{ImageId: aws.String("3")},
									},
								},
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("4")},
										{ImageId: aws.String("5")},
										{ImageId: aws.String("6")},
									},
								},
							},
						}, false)
						callback(&ec2.DescribeInstancesOutput{
							Reservations: []*ec2.Reservation{
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("7")},
										{ImageId: aws.String("8")},
										{ImageId: aws.String("9")},
									},
								},
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("10")},
										{ImageId: aws.String("11")},
										{ImageId: aws.String("12")},
									},
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Instance{
				{ImageId: aws.String("1")},
				{ImageId: aws.String("2")},
				{ImageId: aws.String("3")},
				{ImageId: aws.String("4")},
				{ImageId: aws.String("5")},
				{ImageId: aws.String("6")},
				{ImageId: aws.String("7")},
				{ImageId: aws.String("8")},
				{ImageId: aws.String("9")},
				{ImageId: aws.String("10")},
				{ImageId: aws.String("11")},
				{ImageId: aws.String("12")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllInstances()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllInstances()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

func Test_ec2Repository_ListAllKeyPairs(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *MockEC2Client)
		want    []*ec2.KeyPairInfo
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *MockEC2Client) {
				client.On("DescribeKeyPairs", &ec2.DescribeKeyPairsInput{}).
					Return(&ec2.DescribeKeyPairsOutput{
						KeyPairs: []*ec2.KeyPairInfo{
							{KeyPairId: aws.String("1")},
							{KeyPairId: aws.String("2")},
							{KeyPairId: aws.String("3")},
							{KeyPairId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []*ec2.KeyPairInfo{
				{KeyPairId: aws.String("1")},
				{KeyPairId: aws.String("2")},
				{KeyPairId: aws.String("3")},
				{KeyPairId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockEC2Client{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  cache.New(5 * 1024),
			}
			got, err := r.ListAllKeyPairs()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			// Check that results were cached
			cachedData, err := r.ListAllKeyPairs()
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, cachedData))
		})
	}
}

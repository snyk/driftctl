package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/stretchr/testify/assert"
)

func TestNewS3ReaderInvalid(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *S3Backend
		wantErr error
	}{
		{
			name: "invalid path",
			args: args{
				path: "foobar",
			},
			want:    nil,
			wantErr: fmt.Errorf("Unable to parse S3 path: foobar. Must be BUCKET_NAME/PATH/TO/OBJECT"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewS3Reader(tt.args.path)
			if err.Error() != tt.wantErr.Error() {
				t.Errorf("NewS3Reader() error = '%s', wantErr '%s'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewS3Reader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewS3Reader(t *testing.T) {
	assert := assert.New(t)
	reader, err := NewS3Reader("sample_bucket/path/to/state.tfstate")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(
		"path/to/state.tfstate",
		*reader.input.Key,
	)
	assert.Equal(
		"sample_bucket",
		*reader.input.Bucket,
	)
}

func TestS3Backend_ReadWithError(t *testing.T) {
	assert := assert.New(t)
	fakeS3 := &mocks.FakeS3{}
	fakeErr := &mocks.FakeRequestFailure{}
	fakeErr.On("Message").Return("Request failed on aws side")
	fakeS3.On("GetObject", mock.Anything).Return(nil, fakeErr)

	reader, err := NewS3Reader("foobar/path/to/state")
	if err != nil {
		t.Error(err)
	}
	reader.S3Client = fakeS3
	var b []byte
	n, err := reader.Read(b)
	assert.Empty(n)
	assert.Equal("Error reading state 'path/to/state' from s3 bucket 'foobar': Request failed on aws side", err.Error())
}

func TestS3Backend_Read(t *testing.T) {
	assert := assert.New(t)
	fakeS3 := &mocks.FakeS3{}
	fakeResponse, _ := os.Open("testdata/valid.tfstate")
	defer fakeResponse.Close()
	fakeS3.On("GetObject", &s3.GetObjectInput{
		Bucket: aws.String("foobar"),
		Key:    aws.String("path/to/state"),
	}).Return(&s3.GetObjectOutput{Body: fakeResponse}, nil).Once()

	reader, err := NewS3Reader("foobar/path/to/state")
	if err != nil {
		t.Error(err)
	}
	reader.S3Client = fakeS3
	_, err = ioutil.ReadAll(reader)
	assert.Nil(err)
}

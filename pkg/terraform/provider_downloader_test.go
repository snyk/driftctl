package terraform

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	terraformError "github.com/snyk/driftctl/pkg/terraform/error"

	"github.com/stretchr/testify/assert"

	"github.com/jarcoal/httpmock"
)

func TestProviderDownloader_Download(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	downloader := NewProviderDownloader()
	url := "https://example.com/terraform-provider-aws_3.19.0_linux_amd64.zip"

	cases := []struct {
		name       string
		httpStatus *int
		testFile   *string
		responder  httpmock.Responder
		assert     func(assert *assert.Assertions, tmpDir string, err error)
	}{
		{
			name:      "TestBadResponse(404)",
			responder: httpmock.NewBytesResponder(http.StatusNotFound, []byte{}),
			assert: func(assert *assert.Assertions, tmpDir string, err error) {
				assert.Equal(
					fmt.Sprintf("unsuccessful request to %s: 404", url),
					err.Error(),
				)
			},
		},
		{
			name:      "TestProviderNotFound(403)",
			responder: httpmock.NewBytesResponder(http.StatusForbidden, []byte{}),
			assert: func(assert *assert.Assertions, tmpDir string, err error) {
				assert.IsType(
					terraformError.ProviderNotFoundError{},
					err,
				)
			},
		},
		{
			name:      "TestHttpError",
			responder: httpmock.NewErrorResponder(fmt.Errorf("test error")),
			assert: func(assert *assert.Assertions, tmpDir string, err error) {
				assert.Contains(err.Error(), "test error")
			},
		},
		{
			name:     "TestInvalidZip",
			testFile: aws.String("invalid.zip"),
			assert: func(assert *assert.Assertions, tmpDir string, err error) {
				assert.NotNil(err)
				infos, err := ioutil.ReadDir(tmpDir)
				assert.Nil(err)
				assert.Len(infos, 0)
			},
		},
		{
			name:     "TestValidZip",
			testFile: aws.String("terraform-provider-aws_3.5.0_linux_amd64.zip"),
			assert: func(assert *assert.Assertions, tmpDir string, err error) {
				assert.Nil(err)
				file, err := ioutil.ReadFile(path.Join(tmpDir, "terraform-provider-aws_v3.5.0_x5"))
				assert.Nil(err)
				assert.Equal([]byte{0x74, 0x65, 0x73, 0x74, 0xa}, file)
			},
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(tt *testing.T) {
			tmpDir := tt.TempDir()

			httpmock.Reset()
			assert := assert.New(tt)

			if c.httpStatus == nil {
				c.httpStatus = aws.Int(http.StatusOK)
			}

			if c.responder != nil {
				httpmock.RegisterResponder("GET", url, c.responder)
			} else {
				if c.testFile != nil {
					body, err := ioutil.ReadFile("./testdata/" + *c.testFile)
					if err != nil {
						tt.Error(err)
					}
					httpmock.RegisterResponder("GET", url, httpmock.NewBytesResponder(*c.httpStatus, body))
				}
			}

			err := downloader.Download(url, tmpDir)

			c.assert(assert, tmpDir, err)
		})

	}
}

package telemetry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/version"
	"github.com/cloudskiff/driftctl/test/resource"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSendTelemetry(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name         string
		analysis     *analyser.Analysis
		expectedBody *telemetry
		response     *http.Response
	}{
		{
			name: "valid analysis",
			analysis: func() *analyser.Analysis {
				a := &analyser.Analysis{}
				a.AddManaged(&resource.FakeResource{})
				a.AddUnmanaged(&resource.FakeResource{})
				return a
			}(),
			expectedBody: &telemetry{
				Version:        version.Current(),
				Os:             runtime.GOOS,
				Arch:           runtime.GOARCH,
				TotalResources: 2,
				TotalManaged:   1,
			},
		},
		{
			name:     "nil analysis",
			analysis: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			if tt.expectedBody != nil {
				httpmock.RegisterResponder(
					"POST",
					"https://2lvzgmrf2e.execute-api.eu-west-3.amazonaws.com/telemetry",
					func(req *http.Request) (*http.Response, error) {

						requestTelemetry := &telemetry{}
						requestBody, err := ioutil.ReadAll(req.Body)
						if err != nil {
							t.Fatal(err)
						}
						err = json.Unmarshal(requestBody, requestTelemetry)
						if err != nil {
							t.Fatal(err)
						}

						assert.Equal(t, tt.expectedBody, requestTelemetry)

						response := tt.response
						if response == nil {
							response = httpmock.NewBytesResponse(202, []byte{})
						}
						return response, nil
					},
				)
			}
			SendTelemetry(tt.analysis)
		})
	}
}

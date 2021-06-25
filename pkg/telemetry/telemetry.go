package telemetry

import (
	"bytes"
	"net/http"

	"github.com/cloudskiff/driftctl/pkg/memstore"
	"github.com/sirupsen/logrus"
)

func SendTelemetry(s memstore.Store) {
	body, err := s.Bucket(memstore.TelemetryBucket).MarshallJSON()
	if err != nil {
		logrus.Debug(err)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://2lvzgmrf2e.execute-api.eu-west-3.amazonaws.com/telemetry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	if err != nil {
		logrus.Debugf("Unable to send telemetry data: %+v", err)
		return
	}

}

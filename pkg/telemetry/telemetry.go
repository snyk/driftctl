package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/cloudskiff/driftctl/pkg/memstore"
	"github.com/sirupsen/logrus"
)

type telemetry struct {
	Version        string `json:"version"`
	Os             string `json:"os"`
	Arch           string `json:"arch"`
	TotalResources int    `json:"total_resources"`
	TotalManaged   int    `json:"total_managed"`
	Duration       uint   `json:"duration"`
}

func SendTelemetry(s memstore.Store) {
	t := &telemetry{}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("version").(string); ok {
		t.Version = val
	}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("os").(string); ok {
		t.Os = val
	}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("arch").(string); ok {
		t.Arch = val
	}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("total_resources").(int); ok {
		t.TotalResources = val
	}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("total_managed").(int); ok {
		t.TotalManaged = val
	}

	if val, ok := s.Bucket(memstore.TelemetryBucket).Get("duration").(uint); ok {
		t.Duration = val
	}

	body, err := json.Marshal(t)
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

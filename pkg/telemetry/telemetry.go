package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/version"
	"github.com/sirupsen/logrus"
)

type telemetry struct {
	Version          string `json:"version"`
	Os               string `json:"os"`
	Arch             string `json:"arch"`
	TotalResources   int    `json:"total_resources"`
	TotalManaged     int    `json:"total_managed"`
	Duration         uint   `json:"duration"`
	IgnoreRulesCount int    `json:"ignore_rules_count"`
}

func SendTelemetry(analysis *analyser.Analysis) {

	if analysis == nil {
		return
	}

	t := telemetry{
		Version:          version.Current(),
		Os:               runtime.GOOS,
		Arch:             runtime.GOARCH,
		TotalResources:   analysis.Summary().TotalResources,
		TotalManaged:     analysis.Summary().TotalManaged,
		Duration:         uint(analysis.Duration.Seconds() + 0.5),
		IgnoreRulesCount: analysis.IgnoreRulesCount,
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

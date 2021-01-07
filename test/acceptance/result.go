package acceptance

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"

	"github.com/stretchr/testify/require"
)

type ScanResult struct {
	*require.Assertions
	analyser.Analysis
}

func NewScanResult(t *testing.T, analysis analyser.Analysis) *ScanResult {
	return &ScanResult{
		Assertions: require.New(t),
		Analysis:   analysis,
	}
}

func (r *ScanResult) AssertResourceUnmanaged(id, ty string) {
	for _, u := range r.Unmanaged() {
		if u.TerraformType() == ty && u.TerraformId() == id {
			return
		}
	}
	r.Failf("Resource not unmanaged", "%s(%s)", id, ty)
}

func (r *ScanResult) AssertResourceDeleted(id, ty string) {
	for _, u := range r.Deleted() {
		if u.TerraformType() == ty && u.TerraformId() == id {
			return
		}
	}
	r.Failf("Resource not deleted", "%s(%s)", id, ty)
}

func (r *ScanResult) AssertResourceDriftCount(id, ty string, count int) {
	for _, u := range r.Differences() {
		if u.Res.TerraformType() == ty && u.Res.TerraformId() == id {
			r.Equal(count, len(u.Changelog))
		}
	}
	r.Failf("no differences found", "%s(%s)", id, ty)
}

func (r *ScanResult) AssertResourceHasDrift(id, ty string, change analyser.Change) {
	found := false
	for _, u := range r.Differences() {
		if u.Res.TerraformType() == ty && u.Res.TerraformId() == id {
			changelogStr, _ := json.MarshalIndent(u.Changelog, "", " ")
			changeStr, _ := json.MarshalIndent(change, "", " ")
			r.Contains(u.Changelog, change, fmt.Sprintf("Change not found\nCHANGE: %s\nCHANGELOG:\n%s", changeStr, changelogStr))
			found = true
		}
	}
	if !found {
		r.Failf("no differences found", "%s(%s)", id, ty)
	}
}

func (r *ScanResult) AssertResourceHasNoDrift(id, ty string) {
	for _, u := range r.Differences() {
		if u.Res.TerraformType() == ty && u.Res.TerraformId() == id {
			changelogStr, _ := json.MarshalIndent(u.Changelog, "", " ")
			r.Failf("resource has drifted", "%s(%s) :\n %v", id, ty, changelogStr)
		}
	}
}

func (r *ScanResult) AssertCoverage(expected int) {
	r.Equal(expected, r.Coverage)
}

func (r *ScanResult) AssertDriftCountTotal(count int) {
	driftCount := 0
	for _, diff := range r.Differences() {
		driftCount += len(diff.Changelog)
	}
	r.Equal(count, driftCount)
}

func (r ScanResult) AssertInfrastructureIsInSync() {
	r.Equal(true, r.Analysis.IsSync(), "Infrastructure is not in sync")
}

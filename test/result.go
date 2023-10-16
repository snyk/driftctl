package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/snyk/driftctl/pkg/analyser"

	"github.com/stretchr/testify/require"
)

type ScanResult struct {
	*require.Assertions
	*analyser.Analysis
}

func NewScanResult(t *testing.T, analysis *analyser.Analysis) *ScanResult {
	return &ScanResult{
		Assertions: require.New(t),
		Analysis:   analysis,
	}
}

func (r *ScanResult) AssertResourceUnmanaged(id, ty string) {
	for _, u := range r.Unmanaged() {
		if u.ResourceType() == ty && u.ResourceId() == id {
			return
		}
	}
	r.Failf("Resource not unmanaged", "%s(%s)", id, ty)
}

func (r *ScanResult) AssertResourceDeleted(id, ty string) {
	for _, u := range r.Deleted() {
		if u.ResourceType() == ty && u.ResourceId() == id {
			return
		}
	}
	r.Failf("Resource not deleted", "%s(%s)", id, ty)
}

func (r *ScanResult) AssertCoverage(expected int) {
	r.Equal(expected, r.Coverage())
}

func (r *ScanResult) AssertDeletedCount(count int) {
	r.Equal(count, len(r.Deleted()))
}

func (r *ScanResult) AssertManagedCount(count int) {
	r.Equal(count, len(r.Managed()))
}

func (r *ScanResult) AssertUnmanagedCount(count int) {
	r.Equal(count, len(r.Unmanaged()))
}

func (r ScanResult) AssertInfrastructureIsInSync() {
	r.Equal(
		true,
		r.Analysis.IsSync(),
		fmt.Sprintf(
			"Infrastructure is not in sync: \n%s\n",
			r.printAnalysisResult(),
		),
	)
}

func (r ScanResult) AssertInfrastructureIsNotSync() {
	r.Equal(
		false,
		r.Analysis.IsSync(),
		fmt.Sprintf(
			"Infrastructure is in sync: \n%s\n",
			r.printAnalysisResult(),
		),
	)
}

func (r *ScanResult) printAnalysisResult() string {
	str, _ := json.MarshalIndent(r.Analysis, "", " ")
	return string(str)
}

package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressTimeoutDoesNotInc(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	progress.Inc()
	progress.Stop() // should not hang
	progress.Inc()  // should restart progress and inc
	assert.Equal(t, uint64(2), progress.Val())
	assert.Equal(t, true, progress.started.Load())

	progress.Stop()
	assert.Equal(t, false, progress.started.Load())
}

func TestProgressTimeoutDoesNotHang(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	time.Sleep(progressTimeout)
	for progress.started.Load() == true {
	}
	progress.Inc()  // should not hang but inc
	progress.Stop() // should not hang
	assert.Equal(t, uint64(1), progress.Val())
}

func TestProgress(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	progress.Inc()
	progress.Inc()
	progress.Inc()
	progress.Stop()
	assert.Equal(t, uint64(3), progress.Val())
}

package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressTimeoutDoesNotInc(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	progress.Stop() // should not hang
	progress.Inc()  // should not hang or inc
	assert.Equal(t, uint64(0), progress.Val())
}

func TestProgressTimeoutDoesNotHang(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	time.Sleep(progressTimeout)
	for progress.started.Load() == true {
	}
	progress.Inc()  // should not hang or inc
	progress.Stop() // should not hang
	assert.Equal(t, uint64(0), progress.Val())

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

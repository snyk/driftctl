package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressTimeout(t *testing.T) {
	progress := NewProgress()
	progress.Start()
	time.Sleep(progressTimeout + 1)
	progress.Inc()  // should not hang
	progress.Stop() // should not hang
	assert.Equal(t, uint64(0), progress.Val())
}

func TestProgress(t *testing.T) {
	progress := NewProgress()
	progress.Start()
	progress.Inc()
	progress.Inc()
	progress.Inc()
	progress.Stop()
	assert.Equal(t, uint64(3), progress.Val())
}

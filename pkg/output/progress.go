package output

import (
	"time"

	"go.uber.org/atomic"

	"github.com/sirupsen/logrus"
)

var spinner = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

const (
	progressTimeout     = 10 * time.Second
	progressRefreshRate = 200 * time.Millisecond
)

type Progress interface {
	Start()
	Stop()
	Inc()
	Val() uint64
}

type progress struct {
	endChan chan struct{}
	started *atomic.Bool
	count   *atomic.Uint64
}

func NewProgress() *progress {
	return &progress{
		nil,
		atomic.NewBool(false),
		atomic.NewUint64(0),
	}
}

func (p *progress) Start() {
	if !p.started.Swap(true) {
		p.count.Store(0)
		p.endChan = make(chan struct{})
		go p.watch()
		go p.render()
	}
}

func (p *progress) Stop() {
	if p.started.Swap(false) {
		Printf("Scanned resources:    (%d)\n", p.count.Load())
		close(p.endChan)
	}
}

func (p *progress) Inc() {
	if lastVal := p.count.Load(); !p.started.Load() {
		logrus.Debug("Progress received a tic after stopping. Restarting...")
		p.Start()
		p.count.Store(lastVal)
	}
	p.count.Inc()
}

func (p *progress) Val() uint64 {
	return p.count.Load()
}

func (p *progress) render() {
	i := -1
	Printf("Scanning resources:\r")
	for {
		select {
		case <-p.endChan:
			return
		case <-time.After(progressRefreshRate):
			i++
			if i >= len(spinner) {
				i = 0
			}
			Printf("Scanning resources: %s (%d)\r", spinner[i], p.count.Load())
		}
	}
}

func (p *progress) watch() {
Loop:
	for {
		lastVal := p.count.Load()
		select {
		case <-time.After(progressTimeout):
			if p.count.Load() != lastVal {
				continue
			}
			if p.started.Swap(false) {
				close(p.endChan)
				break Loop
			}
		case <-p.endChan:
			return
		}
	}
	logrus.Debug("Progress did not receive any tic. Stopping...")
}

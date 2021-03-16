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
	ticChan chan struct{}
	endChan chan struct{}
	started *atomic.Bool
	count   *atomic.Uint64
}

func NewProgress() *progress {
	return &progress{
		make(chan struct{}),
		make(chan struct{}),
		atomic.NewBool(false),
		atomic.NewUint64(0),
	}
}

func (p *progress) Start() {
	if !p.started.Swap(true) {
		go p.watch()
		go p.render()
	}
}

func (p *progress) Stop() {
	if p.started.Swap(false) {
		p.endChan <- struct{}{}
		Printf("\n")
	}
}

func (p *progress) Inc() {
	if p.started.Load() {
		p.ticChan <- struct{}{}
	}
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
		select {
		case <-p.ticChan:
			p.count.Inc()
			continue Loop
		case <-time.After(progressTimeout):
			p.started.Store(false)
			break Loop
		case <-p.endChan:
			return
		}
	}
	logrus.Debug("Progress did not receive any tic. Stopping...")
	p.endChan <- struct{}{}
}

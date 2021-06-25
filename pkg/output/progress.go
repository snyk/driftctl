package output

import (
	"fmt"
	"time"

	"go.uber.org/atomic"

	"github.com/sirupsen/logrus"
)

var spinner = []string{"⣷", "⣯", "⣟", "⡿", "⢿", "⣻", "⣽", "⣾"}

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

type ProgressOptions struct {
	LoadingText  string
	FinishedText string
	ShowCount    bool
}

type progress struct {
	endChan           chan struct{}
	started           *atomic.Bool
	count             *atomic.Uint64
	loadingText       string
	finishedText      string
	showCount         bool
	highestLineLength int
}

func NewProgress(loadingText, finishedText string, showCount bool) *progress {
	return &progress{
		nil,
		atomic.NewBool(false),
		atomic.NewUint64(0),
		loadingText,
		finishedText,
		showCount,
		0,
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
		if p.showCount {
			p.printf("%s (%d)\n", p.finishedText, p.count.Load())
		} else {
			p.printf("%s\r", p.finishedText)
		}
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
	p.printf("%s\r", p.loadingText)
	for {
		select {
		case <-p.endChan:
			return
		case <-time.After(progressRefreshRate):
			i++
			if i >= len(spinner) {
				i = 0
			}
			if p.showCount {
				p.printf("%s %s (%d)\r", p.loadingText, spinner[i], p.count.Load())
			} else {
				p.printf("%s %s\r", p.loadingText, spinner[i])
			}
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

func (p *progress) flush() {
	for i := 0; i < p.highestLineLength; i++ {
		Printf(" ")
	}
	Printf("\r")
}

func (p *progress) printf(format string, args ...interface{}) {
	txt := fmt.Sprintf(format, args...)
	length := len(txt)
	if length > p.highestLineLength {
		p.highestLineLength = length
	}
	p.flush()
	Printf(txt)
}

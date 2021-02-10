package parallel

import (
	"context"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"go.uber.org/atomic"

	"golang.org/x/sync/semaphore"
)

type ParallelRunner struct {
	sem     *semaphore.Weighted
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	resChan chan interface{}
	err     error
	hasErr  *atomic.Bool
	waiting *atomic.Bool
}

func NewParallelRunner(ctx context.Context, maxRun int64) *ParallelRunner {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &ParallelRunner{
		sem:     semaphore.NewWeighted(maxRun),
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancelFunc,
		resChan: make(chan interface{}),
		err:     nil,
		hasErr:  atomic.NewBool(false),
		waiting: atomic.NewBool(false),
	}
}

func (p *ParallelRunner) SubRunner() *ParallelRunner {
	ctx, cancelFunc := context.WithCancel(p.ctx)
	return &ParallelRunner{
		sem:     p.sem,
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancelFunc,
		resChan: make(chan interface{}),
		err:     nil,
		hasErr:  atomic.NewBool(false),
		waiting: atomic.NewBool(false),
	}
}

func (p *ParallelRunner) Read() chan interface{} {
	p.wait()
	return p.resChan
}

func (p *ParallelRunner) DoneChan() <-chan struct{} {
	return p.ctx.Done()
}

func (p *ParallelRunner) Err() error {
	return p.err
}

func (p *ParallelRunner) wait() {
	if !p.waiting.Swap(true) {
		go func() {
			p.wg.Wait()
			close(p.resChan)
		}()
	}
}

func (p *ParallelRunner) Run(runnable func() (interface{}, error)) {
	p.wg.Add(1)
	go func() {
		if err := p.sem.Acquire(p.ctx, 1); err == nil {
			// only release if sem was acquired
			defer p.sem.Release(1)
		}
		defer p.wg.Done()
		// Prevent new routines executions if we already got an error from another routine
		if p.ctx.Err() != nil {
			return
		}
		// Handle panic in routines and stop runner with proper error
		// Some failed call to grpc plugin like getSchema trigger a panic
		defer func() {
			if r := recover(); r != nil {
				sentry.CurrentHub().Recover(r)
				p.Stop(errors.Errorf("A runner routine paniced: %s", r))
			}
		}()
		res, err := runnable()
		if err != nil {
			p.Stop(err)
		}
		p.resChan <- res
	}()
}

func (p *ParallelRunner) Stop(err error) {
	if !p.hasErr.Swap(true) {
		logrus.Debug("Stopping ParallelRunner")
		p.err = err
		p.cancel()
	}
}

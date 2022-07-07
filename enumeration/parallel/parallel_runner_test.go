package parallel

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestParallelRunner_RunOk(t *testing.T) {
	assert := assert.New(t)

	runner := NewParallelRunner(context.TODO(), 10)

	for i := 0; i < 100; i++ {
		runner.Run(func() (interface{}, error) {
			return 1, nil
		})
	}

	val := 0
Loop:
	for {
		select {
		case res, ok := <-runner.Read():
			if !ok {
				break Loop
			}
			val += res.(int)
		case <-runner.DoneChan():
			break Loop
		}
	}

	assert.Nil(runner.Err())
	assert.Equal(100, val)
}

func TestParallelRunner_RunWithError(t *testing.T) {
	assert := assert.New(t)

	runner := NewParallelRunner(context.TODO(), 10)

	err := errors.New("test")

	for i := 0; i < 100; i++ {
		val := i
		runner.Run(func() (interface{}, error) {
			if val <= 10 {
				return 0, err
			}
			return 1, nil
		})
	}

	val := 0
Loop:
	for {
		select {
		case res, ok := <-runner.Read():
			if !ok {
				break Loop
			}
			val += res.(int)
		case <-runner.DoneChan():
			break Loop
		}
	}

	assert.Equal(err, runner.Err())
	assert.Less(val, 100)
}

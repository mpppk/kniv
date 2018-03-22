package kniv

import (
	"errors"
	"time"
)

type DelayProcessor struct {
	*BaseProcessor
	sleepMilliSec time.Duration
}

type DelayProcessorArgs struct {
	*BaseArgs
	IntervalMilliSec time.Duration
	Group            string
}

func NewDelayProcessor(args *DelayProcessorArgs) *DelayProcessor {
	delayProcessor := &DelayProcessor{
		BaseProcessor: &BaseProcessor{
			Name:   "delay processor",
			inChan: make(chan Event, args.QueueSize),
		},
		sleepMilliSec: args.IntervalMilliSec * time.Millisecond,
	}
	delayProcessor.BaseProcessor.Process = delayProcessor.wait
	return delayProcessor
}

func NewDelayProcessorFromArgs(intfArgs interface{}) (*DelayProcessor, error) {
	args, ok := intfArgs.(DelayProcessorArgs)
	if !ok {
		return nil, errors.New("invalid delay processor args")
	}
	return NewDelayProcessor(&args), nil
}

func (d *DelayProcessor) wait(event Event) ([]Event, error) {
	time.Sleep(d.sleepMilliSec)
	return []Event{event}, nil
}

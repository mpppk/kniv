package kniv

import (
	"errors"
	"fmt"
	"time"
)

const delayProcessorName = "delay"

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
			Name:   delayProcessorName,
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

type DelayProcessorGenerator struct{}

func (g *DelayProcessorGenerator) GetName() string {
	return delayProcessorName
}

func (g *DelayProcessorGenerator) Generate(intfArgs interface{}) (Processor, error) {
	args, ok := intfArgs.(DelayProcessorArgs)
	if !ok {
		return nil, fmt.Errorf("invalid delay processor args: %#v", intfArgs)
	}
	return NewDelayProcessor(&args), nil
}

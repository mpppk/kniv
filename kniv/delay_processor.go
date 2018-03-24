package kniv

import (
	"fmt"
	"time"
)

const delayProcessorType = "delay"

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
			Type:   delayProcessorType,
			Name:   delayProcessorType,
			inChan: make(chan Event, args.QueueSize),
		},
		sleepMilliSec: args.IntervalMilliSec * time.Millisecond,
	}
	delayProcessor.BaseProcessor.Process = delayProcessor.wait
	return delayProcessor
}

func (d *DelayProcessor) wait(event Event) ([]Event, error) {
	time.Sleep(d.sleepMilliSec)
	return []Event{event}, nil
}

type DelayProcessorGenerator struct{}

func (g *DelayProcessorGenerator) Generate(intfArgs interface{}) (Processor, error) {
	args, ok := intfArgs.(DelayProcessorArgs)
	if !ok {
		return nil, fmt.Errorf("invalid delay processor args: %#v", intfArgs)
	}
	return NewDelayProcessor(&args), nil
}

func (g *DelayProcessorGenerator) GetType() string {
	return delayProcessorType
}

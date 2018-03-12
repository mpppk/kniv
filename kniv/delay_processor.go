package kniv

import "time"

type DelayProcessor struct {
	*BaseProcessor
	sleepMilliSec time.Duration
}

func NewDelayProcessor(queueSize int, sleepMilliSec time.Duration) *DelayProcessor {
	delayProcessor := &DelayProcessor{
		BaseProcessor: &BaseProcessor{
			Name:   "delay processor",
			inChan: make(chan Event, queueSize),
		},
		sleepMilliSec: sleepMilliSec * time.Millisecond,
	}
	delayProcessor.BaseProcessor.Process = delayProcessor.wait
	return delayProcessor
}

func (d *DelayProcessor) wait(resource Event) ([]Event, error) {
	time.Sleep(d.sleepMilliSec)
	resource.PushLabel("twitter.image") // FIXME temp
	return []Event{resource}, nil
}
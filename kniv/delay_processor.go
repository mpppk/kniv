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
			inChan: make(chan Resource, queueSize),
		},
		sleepMilliSec: sleepMilliSec * time.Millisecond,
	}
	delayProcessor.BaseProcessor.Process = delayProcessor.wait
	return delayProcessor
}

func (d *DelayProcessor) wait(resource Resource) ([]Resource, error) {
	time.Sleep(d.sleepMilliSec)
	resource.ResourceType = "twitter.image" // FIXME temp
	resource.NextResourceType = "end"       // FIXME temp
	return []Resource{resource}, nil
}

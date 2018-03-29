package kniv

import (
	"fmt"
	"log"
)

type processors []Processor

func (ps processors) get(name string) (Processor, bool) {
	for _, processor := range ps {
		if processor.GetName() == name {
			return processor, true
		}
	}
	return nil, false
}

func (ps processors) getOrCreate(pType string) (Processor, error) {
	for _, processor := range ps {
		if processor.GetType() == pType {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("%s not found", pType)
}

func RegisterProcessorsFromFlow(dispatcher *Dispatcher, flow *Flow, factory ProcessorFactory) error {
	// FIXME return Job with processor struct and register outside

	var ps processors
	for _, pipeline := range flow.Pipelines {
		log.Printf("Pipline [%s] loading... ", pipeline.Name)
		for i, job := range pipeline.Jobs {
			name := job.GetProcessorType() // FIXME check job processorName if exist
			log.Printf("Job [%s] loading... ", name)

			var newProcessor Processor
			if p, ok := ps.get(name); ok {
				newProcessor = p
			} else {
				processor, err := factory.Create(job)
				if err != nil {
					return err
				}
				newProcessor = processor
			}

			var fullConsumeLabels []Label
			var fullProduceLabels []Label

			if len(job.Consume) == 0 {
				label := Label(job.ProcessorType)
				if job.Name != "" {
					label = Label(job.Name)
				}
				fullConsumeLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, label))
				fullConsumeLabels = append(fullConsumeLabels, fullConsumeLabel)
				if i == 0 {
					fullConsumeLabels = append(fullConsumeLabels, Label(pipeline.Name))
				}
			} else {
				for _, c := range job.Consume {
					if c == "init" { // FIXME
						fullConsumeLabels = append(fullConsumeLabels, c)
						continue
					}

					if _, ok := flow.getPipeline(string(c)); ok {
						fullConsumeLabels = append(fullConsumeLabels, c)
						continue
					}

					fullConsumeLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, c))
					fullConsumeLabels = append(fullConsumeLabels, fullConsumeLabel)
				}
			}

			if len(job.Produce) == 0 && i < (len(pipeline.Jobs)-1) {
				nextJob := pipeline.Jobs[i+1]
				label := Label(nextJob.ProcessorType)
				if nextJob.Name != "" {
					label = Label(nextJob.Name)
				}
				fullProduceLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, label))
				fullProduceLabels = append(fullProduceLabels, fullProduceLabel)
			} else {
				for _, p := range job.Produce {
					if _, ok := flow.getPipeline(string(p)); ok {
						fullProduceLabels = append(fullProduceLabels, p)
						continue
					}

					fullProduceLabel := Label(fmt.Sprintf("%s/%s", pipeline.Name, p))
					fullProduceLabels = append(fullProduceLabels, fullProduceLabel)
				}
			}

			dispatcher.RegisterTask(newProcessor.GetName(), fullConsumeLabels, fullProduceLabels, newProcessor) // FIXME name
			log.Printf("new task is registered: %s, consumes: %s, produces: %s", newProcessor.GetName(), fullConsumeLabels, fullProduceLabels)

		}
	}
	return nil
}

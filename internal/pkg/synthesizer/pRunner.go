package synthesizer

import "sync"

//PartProcessor interface
type PartProcessor interface {
	Process(*TTSDataPart) error
}

//PartRunner runs parts of the job
type PartRunner struct {
	processors []PartProcessor
}

//NewPartRunner creates parallel runner
func NewPartRunner() *PartRunner {
	return &PartRunner{}
}

//Process is main method
func (p *PartRunner) Process(data *TTSData) error {
	workerQueueLimit := make(chan bool, 5)
	var wg sync.WaitGroup

	var errorMain error
	for _, part := range data.Parts {
		workerQueueLimit <- true // try get access to work
		wg.Add(1)
		go func(part *TTSDataPart) {
			defer wg.Done()
			defer func() { <-workerQueueLimit }()
			err := p.process(part)
			if err != nil {
				errorMain = err
			}
		}(part)
	}
	wg.Wait()
	if errorMain != nil {
		return errorMain
	}
	return nil
}

//Add adds a processor to the end
func (p *PartRunner) Add(pr PartProcessor) {
	p.processors = append(p.processors, pr)
}

func (p *PartRunner) process(data *TTSDataPart) error {
	for _, pr := range p.processors {
		err := pr.Process(data)
		if err != nil {
			return err
		}
	}
	return nil
}

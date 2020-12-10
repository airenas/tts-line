package synthesizer

import (
	"sync"

	"github.com/pkg/errors"
)

//PartProcessor interface
type PartProcessor interface {
	Process(*TTSDataPart) error
}

//PartRunner runs parts of the job
type PartRunner struct {
	processors     []PartProcessor
	parallelWorker int
}

//NewPartRunner creates parallel runner
func NewPartRunner(parallelWorker int) *PartRunner {
	if parallelWorker < 1 {
		parallelWorker = 3
	}
	return &PartRunner{parallelWorker: parallelWorker}
}

//Process is main method
func (p *PartRunner) Process(data *TTSData) error {
	workerQueueLimit := make(chan bool, p.parallelWorker)
	errCh := make(chan error, 1)
	closeCh := make(chan bool, 1)
	var wg sync.WaitGroup

	defer close(closeCh)

	for _, part := range data.Parts {
		select {
		case err := <-errCh:
			return errors.Wrap(err, "Failed to process partial tasks")
		case workerQueueLimit <- true:
			wg.Add(1)
			go func(part *TTSDataPart) {
				defer wg.Done()
				defer func() { <-workerQueueLimit }()
				err := p.process(part, closeCh)
				if err != nil {
					select {
					case <-closeCh:
					case errCh <- err:
					}
				}
			}(part)
		}
	}

	waitCh := make(chan bool, 1)
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case err := <-errCh:
		return errors.Wrap(err, "Failed to process partial tasks")
	case <-waitCh:
	}
	return nil
}

//Add adds a processor to the end
func (p *PartRunner) Add(pr PartProcessor) {
	p.processors = append(p.processors, pr)
}

func (p *PartRunner) process(data *TTSDataPart, clCh <-chan bool) error {
	for _, pr := range p.processors {
		select {
		case <-clCh:
			return errors.New("Unexpected work termination")
		default:
			err := pr.Process(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

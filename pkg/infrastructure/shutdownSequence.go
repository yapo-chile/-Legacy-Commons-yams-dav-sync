package infrastructure

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
)

// ShutdownSequence is a stack implementation to control the shutdown order of each
// infrastructure component needed. First process that starts is the last to be turned off
type ShutdownSequence struct {
	sequence  []io.Closer
	waitGroup *sync.WaitGroup
}

// Push pushes a new component into the stack to be turned off.
func (s *ShutdownSequence) Push(task io.Closer) {
	s.sequence = append(s.sequence, task)
	s.waitGroup.Add(1)
}

// Pop removes the last component pushed
func (s *ShutdownSequence) pop() io.Closer {
	if len(s.sequence) > 0 {
		task := s.sequence[len(s.sequence)-1]
		s.sequence = s.sequence[:len(s.sequence)-1]
		return task
	}
	return nil
}

// NewShutdownSequence creates a new ShutdownSequence
func NewShutdownSequence() *ShutdownSequence {
	var sequence []io.Closer
	var waitGroup sync.WaitGroup
	return &ShutdownSequence{
		sequence:  sequence,
		waitGroup: &waitGroup,
	}
}

// Wait waits until the internal waitGroup counter is zero.
func (s *ShutdownSequence) Wait() {
	s.waitGroup.Wait()
}

// Listen launches a go routines that waits for sigint and then stops each task in the stack.
// You need to call Listen before calling Wait, otherwise you risk waiting indefinitely
func (s *ShutdownSequence) Listen() {
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint)
		<-sigint
		// We received an interrupt signal, shut down.
		s.Done()
	}()
}

// Done finishes every goroutine waiting to be done
func (s *ShutdownSequence) Done() {
	go func() {
		for i := 0; i <= len(s.sequence); i++ {
			if task := s.pop(); task != nil {
				if err := task.Close(); err != nil {
					fmt.Printf("Error closing the task of type %T: %+v\n", task, err)
				}
				s.waitGroup.Done()
			}
		}
		// At this point all processes must be done
		fmt.Printf("\nProceeding to shutdown...\n")
	}()
}

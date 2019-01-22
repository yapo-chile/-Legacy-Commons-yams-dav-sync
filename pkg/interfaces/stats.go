package interfaces

// Stats holds sync process stats
type Stats struct {
	Sent       chan int
	Errors     chan int
	Duplicated chan int
	Processed  chan int
	Skipped    chan int
	NotFound   chan int
	Recovered  chan int
}

// NewStats returns a new instance of Stats
func NewStats() Stats {
	processed := make(chan int, 1)
	skipped := make(chan int, 1)
	notFound := make(chan int, 1)
	sent := make(chan int, 1)
	duplicated := make(chan int, 1)
	errors := make(chan int, 1)
	recovered := make(chan int, 1)

	processed <- 0
	skipped <- 0
	notFound <- 0
	sent <- 0
	errors <- 0
	duplicated <- 0
	recovered <- 0

	return Stats{
		Sent:       sent,
		Processed:  processed,
		Errors:     errors,
		Duplicated: duplicated,
		Skipped:    skipped,
		NotFound:   notFound,
		Recovered:  recovered,
	}
}

// inc increments a a given int var, this is useful to increment channel values
var inc = func(i int) int { return i + 1 }

// Close closes all go channels open by stats struct
func (s *Stats) Close() error {
	close(s.Sent)
	close(s.Processed)
	close(s.Errors)
	close(s.Duplicated)
	close(s.Skipped)
	close(s.NotFound)
	close(s.Recovered)
	return nil
}

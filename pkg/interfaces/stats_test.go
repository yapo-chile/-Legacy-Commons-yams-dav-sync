package interfaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStats(t *testing.T) {
	processed := make(chan int, 1)
	skipped := make(chan int, 1)
	notFound := make(chan int, 1)
	sent := make(chan int, 1)
	duplicated := make(chan int, 1)
	errors := make(chan int, 1)

	processed <- 0
	skipped <- 0
	notFound <- 0
	sent <- 0
	errors <- 0
	duplicated <- 0

	expected := Stats{
		Sent:       sent,
		Processed:  processed,
		Errors:     errors,
		Duplicated: duplicated,
		Skipped:    skipped,
		NotFound:   notFound,
	}

	result := NewStats()
	assert.ObjectsAreEqualValues(expected, result)
}

func TestCloseChannels(t *testing.T) {
	stats := NewStats()
	stats.Close()
	isClosed := func(ch <-chan int) bool {
		select {
		case <-ch:
			return true
		default:
		}
		return false
	}

	assert.True(t, isClosed(stats.Sent))
	assert.True(t, isClosed(stats.Processed))
	assert.True(t, isClosed(stats.Errors))
	assert.True(t, isClosed(stats.Duplicated))
	assert.True(t, isClosed(stats.Skipped))
	assert.True(t, isClosed(stats.NotFound))
}

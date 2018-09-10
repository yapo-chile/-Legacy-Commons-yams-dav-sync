package interfaces

import (
	"testing"
)

type TestLogger struct {
	t *testing.T
}

func (logger TestLogger) Debug(msg string) { logger.t.Log("Debug:", msg) }
func (logger TestLogger) Info(msg string)  { logger.t.Log("Info:", msg) }
func (logger TestLogger) Warn(msg string)  { logger.t.Log("Warn:", msg) }
func (logger TestLogger) Error(msg string) { logger.t.Log("Error:", msg) }

func TestGetImages(t *testing.T) {
	logger := TestLogger{t: t}
	localRepo := LocalRepo{
		Path:   "../..",
		Logger: logger,
	}

	images := localRepo.GetImages()
	for _, image := range images {
		logger.Debug("Image: " + image.FilePath)
	}
}

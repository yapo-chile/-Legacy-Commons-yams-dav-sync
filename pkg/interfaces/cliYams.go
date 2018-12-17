package interfaces

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// CLIYams is a yams client that executes operation on yams repository
type CLIYams struct {
	Interactor usecases.SyncInteractor
	Logger     CLIYamsLogger
}

// CLIYamsLogger logs CLI yams events
type CLIYamsLogger interface {
	LogImage(int, usecases.YamsObject)
}

var layout = "20060102T150405"

// retryPreviousFailedUploads gets images from errorControlRepository and try
// to upload those images to yams one more time. If fails increase the counter of errors
// in repo. Repository only returns images with less than a specific number of errors.
func (handler *CLIYams) retryPreviousFailedUploads(threads, maxErrorQty int) error {
	maxConcurrency := handler.Interactor.YamsRepo.GetMaxConcurrentConns()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup
	for w := 0; w < threads; w++ {
		go handler.sendWorker(w, jobs, &waitGroup, domain.SWRetry)
	}
	handler.Interactor.SyncErrorRepo.SetMaxErrorQty(maxErrorQty)
	nPages := handler.Interactor.SyncErrorRepo.GetPagesQty()
	for pagination := 1; pagination <= nPages; pagination++ {
		result, err := handler.Interactor.SyncErrorRepo.GetErrorSync(pagination)
		if err != nil {
			return err
		}
		for _, imagePath := range result {
			image, err := handler.Interactor.LocalRepo.GetImage(imagePath)
			if err != nil {
				continue
			}
			jobs <- image
		}
	}

	close(jobs)
	waitGroup.Wait()
	return nil
}

// Sync synchronizes images between local repository and yams repository
// using go concurrency
func (handler *CLIYams) Sync(threads, maxErrorQty int, imagesDumpYamsPath string) error {
	maxConcurrency := handler.Interactor.YamsRepo.GetMaxConcurrentConns()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	handler.retryPreviousFailedUploads(threads, maxErrorQty)

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go handler.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}

	// Get the data file with list of images to upload
	file, err := os.Open(imagesDumpYamsPath)
	defer file.Close()

	if err != nil {
		return err
	}

	lastSyncDate := handler.Interactor.LastSyncRepo.GetLastSync()
	scanner := bufio.NewScanner(file)
	var imagePath, imageDateStr string

	// for each image read from file
	for scanner.Scan() {
		tuple := strings.Split(scanner.Text(), " ")
		if !validateTuple(tuple, lastSyncDate) {
			continue
		}
		imageDateStr = tuple[0]
		imagePath = tuple[1]
		image, err := handler.Interactor.LocalRepo.GetImage(imagePath)
		if err != nil {
			continue
		}
		jobs <- image
	}

	// If scanner stops because error
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading data from file: %+v", err)
	}

	handler.Interactor.LastSyncRepo.SetLastSync(imageDateStr)

	close(jobs)
	waitGroup.Wait()

	return nil
}

// validateTuple validates a given tuple string is format []string{dateStr,path}
// and the date after or before of a given date
func validateTuple(tuple []string, date time.Time) bool {
	if len(tuple) != 2 {
		return false
	}
	imageDateStr := tuple[0]
	imageDate, err := time.Parse(layout, imageDateStr)
	if err != nil {
		return false
	}
	if imageDate.After(date) || imageDate.Equal(date) {
		return true
	}
	return false
}

// List prints a list of available images in yams repository
func (handler *CLIYams) List() error {
	list, err := handler.Interactor.List()
	for i, img := range list {
		handler.Logger.LogImage(i+1, img)
	}
	return err
}

// Delete deletes an object in yams repository
func (handler *CLIYams) Delete(imageName string) error {
	return handler.Interactor.Delete(imageName)
}

// DeleteAll deletes every imagen in yams repository and redis using concurency
func (handler *CLIYams) DeleteAll(threads int) error {
	images, _ := handler.Interactor.YamsRepo.GetImages()

	jobs := make(chan string)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go handler.deleteWorker(w, jobs, &waitGroup)
	}

	for _, image := range images {
		jobs <- image.ID
	}

	close(jobs)
	waitGroup.Wait()

	return nil
}

// sendWorker sends every image to yams repository
func (handler *CLIYams) sendWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup, previousUploadFailed int) {
	wg.Add(1)
	defer wg.Done()
	for image := range jobs {
		err := handler.Interactor.Send(image)
		if err == nil && previousUploadFailed == domain.SWRetry {
			handler.Interactor.SyncErrorRepo.DelErrorSync(image.Metadata.ImageName)
		}
		if err != nil {
			if err == usecases.ErrYamsDuplicate {
				externalChecksum, _ := handler.Interactor.YamsRepo.HeadImage(image.Metadata.ImageName)
				// If the external image is not updated
				if externalChecksum != image.Metadata.Checksum {
					// delete from yams
					handler.Interactor.YamsRepo.DeleteImage(image.Metadata.ImageName, true)
					// mark to upload in the next sync process (because yams cache)
					handler.Interactor.SyncErrorRepo.SetErrorCounter(image.Metadata.ImageName, 0)
				} else {
					handler.Interactor.SyncErrorRepo.DelErrorSync(image.Metadata.ImageName)
				}
			} else {
				// any other kind of error, mark to upload again in the next sync
				handler.Interactor.SyncErrorRepo.AddErrorSync(image.Metadata.ImageName)
			}
		}
	}
}

// deleteWorker deletes every image to yams repository
func (handler *CLIYams) deleteWorker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	wg.Add(1)
	for j := range jobs {
		handler.Interactor.Delete(j)
	}
	wg.Done()
}

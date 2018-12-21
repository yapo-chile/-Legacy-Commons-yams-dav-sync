package interfaces

import (
	"bufio"
	"fmt"
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
	LogErrorCleaningMarks(imgName string, err error)
	LogErrorRemoteDelete(imgName string, err error)
	LogErrorResetingErrorCounter(imgName string, err error)
	LogErrorIncreasingErrorCounter(imgName string, err error)
}

var layout = "20060102T150405"

// retryPreviousFailedUploads gets images from errorControlRepository and try
// to upload those images to yams one more time. If fails increase the counter of errors
// in repo. Repository only returns images with less than a specific number of errors.
func (cli *CLIYams) retryPreviousFailedUploads(threads, maxErrorTolerance int) {
	maxConcurrency := cli.Interactor.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup
	for w := 0; w < threads; w++ {
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWRetry)
	}
	nPages := cli.Interactor.GetErrorsPagesQty(maxErrorTolerance)
	for pagination := 1; pagination <= nPages; pagination++ {
		result, err := cli.Interactor.GetPreviousErrors(pagination, maxErrorTolerance)
		if err != nil {
			continue
		}
		for _, imagePath := range result {
			image, err := cli.Interactor.GetLocalImage(imagePath)
			if err != nil {
				continue
			}
			jobs <- image
		}
	}

	close(jobs)
	waitGroup.Wait()
}

// Sync synchronizes images between local repository and yams repository
// using go concurrency
func (cli *CLIYams) Sync(threads, maxErrorQty int, imagesDumpYamsPath string) error {
	maxConcurrency := cli.Interactor.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	cli.retryPreviousFailedUploads(threads, maxErrorQty)

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}

	// Get the data file with list of images to upload
	file, e := cli.Interactor.Open(imagesDumpYamsPath)
	if e != nil {
		return e
	}
	defer file.Close() // nolint

	latestSynchronizedImageDate := cli.Interactor.GetLastSynchornizationMark()
	scanner := bufio.NewScanner(file)
	var imagePath, imageDateStr string

	// for each image read from file
	for scanner.Scan() {
		tuple := strings.Split(scanner.Text(), " ")
		if !validateTuple(tuple, latestSynchronizedImageDate) {
			continue
		}
		imageDateStr = tuple[0]
		imagePath = tuple[1]
		image, err := cli.Interactor.GetLocalImage(imagePath)
		if err != nil {
			continue
		}
		jobs <- image
	}

	close(jobs)
	waitGroup.Wait()

	// If scanner stops because error
	if e := scanner.Err(); e != nil {
		return fmt.Errorf("Error reading data from file: %+v", e)
	}
	err := cli.Interactor.SetLastSynchornizationMark(imageDateStr)
	if err != nil {
		return fmt.Errorf("Error setting synchronization mark %+v", err)
	}

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
func (cli *CLIYams) List() error {
	list, err := cli.Interactor.List()
	for i, img := range list {
		cli.Logger.LogImage(i+1, img)
	}
	return err
}

// Delete deletes an object in yams repository
func (cli *CLIYams) Delete(imageName string) error {
	return cli.Interactor.RemoteDelete(imageName)
}

// DeleteAll deletes every imagen in yams repository and redis using concurency
func (cli *CLIYams) DeleteAll(threads int) error {
	images, err := cli.Interactor.List()
	if err != nil {
		return err
	}

	jobs := make(chan string)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		go cli.deleteWorker(w, jobs, &waitGroup)
	}

	for _, image := range images {
		jobs <- image.ID
	}

	close(jobs)
	waitGroup.Wait()

	return nil
}

// sendWorker sends every image to yams repository
func (cli *CLIYams) sendWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup, previousUploadFailed int) {
	wg.Add(1)
	defer wg.Done()
	for image := range jobs {
		imageName := image.Metadata.ImageName
		err := cli.Interactor.Send(image)
		cli.sendErrorControl(
			imageName,
			image.Metadata.Checksum,
			previousUploadFailed,
			err,
		)
	}
}

// sendErrorControl takes action depending of error type retuned by send method
func (cli *CLIYams) sendErrorControl(imageName, imageChecksum string, previousUploadFailed int, err error) {
	switch err {
	case nil:
		if previousUploadFailed == domain.SWRetry {
			if e := cli.Interactor.CleanErrorMarks(imageName); e != nil {
				cli.Logger.LogErrorCleaningMarks(imageName, e)
			}
		}
	case usecases.ErrYamsDuplicate:
		if e := cli.Interactor.RemoteDelete(imageName); e != nil {
			cli.Logger.LogErrorRemoteDelete(imageName, e)
			return
		}
		// mark to upload in the next sync process (because yams cache)
		if e := cli.Interactor.ResetErrorCounter(imageName); e != nil {
			cli.Logger.LogErrorResetingErrorCounter(imageName, e)
		}
	default:
		if e := cli.Interactor.IncreaseErrorCounter(imageName); e != nil {
			cli.Logger.LogErrorIncreasingErrorCounter(imageName, e)
		}
	}
}

// deleteWorker deletes every image to yams repository
func (cli *CLIYams) deleteWorker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	wg.Add(1)
	for j := range jobs {
		if e := cli.Interactor.RemoteDelete(j); e != nil {
			cli.Logger.LogErrorRemoteDelete(j, e)
		}
	}
	wg.Done()
}

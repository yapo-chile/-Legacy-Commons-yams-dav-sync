package interfaces

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// CLIYams is a yams client that executes operation on yams repository
type CLIYams struct {
	yamsService  YamsService
	errorControl ErrorControl
	lastSync     LastSync
	localImage   LocalImage
	logger       CLIYamsLogger
	dateLayout   string
}

// NewCLIYams creates a new instance of CLIYams
func NewCLIYams(yamsService YamsService, errorControl ErrorControl, lastSync LastSync,
	localImage LocalImage, logger CLIYamsLogger, dateLayout string) *CLIYams {
	return &CLIYams{
		yamsService:  yamsService,
		errorControl: errorControl,
		lastSync:     lastSync,
		localImage:   localImage,
		logger:       logger,
		dateLayout:   dateLayout,
	}
}

// YamsService allows operations between local repository & remote yams repository
type YamsService interface {
	// GetRemoteChecksum gets the checksum of image in YAMS
	GetRemoteChecksum(imageName string) (string, *usecases.YamsRepositoryError)
	// Send sends images from local storage to yams bucket
	Send(image domain.Image) *usecases.YamsRepositoryError
	// List gets list of available images in yams bucket
	List() ([]usecases.YamsObject, *usecases.YamsRepositoryError)
	// RemoteDelete deletes image from yams bucket
	RemoteDelete(imageName string, force bool) *usecases.YamsRepositoryError
	// GetMaxConcurrency gets maximum supported concurrency by yams
	GetMaxConcurrency() int
}

// ErrorControl allows operations to control errors with yams synchronization
type ErrorControl interface {
	// GetErrorsPagesQty gets the number of pages for error pagination
	GetErrorsPagesQty(maxErrorTolerance int) int
	// GetPreviousErrors gets a list with previus errors, errors must have its own counter
	// over maxErrorTolerance
	GetPreviousErrors(pagination, maxErrorTolerance int) ([]string, error)
	// CleanErrorMarks cleans every error mark associated with the image
	CleanErrorMarks(imgName string) error
	// SetErrorCounter sets the error counter
	SetErrorCounter(imageName string, counter int) error
	// IncreaseErrorCounter increase the error counter in one, if the image does not
	// have error mark, the mark will be created
	IncreaseErrorCounter(imageName string) error
}

// LastSync allows operations to control latest synchornization status
type LastSync interface {
	// GetLastSynchronizationMark gets the date of latest synchronizated image
	GetLastSynchronizationMark() time.Time
	// SetLastSynchronizationMark sets the date of latest synchronizated image
	SetLastSynchronizationMark(imageDateStr string) error
}

// LocalImage allows operations over local storage
type LocalImage interface {
	// GetLocalImage gets image form local storage parsed as domain.Image
	GetLocalImage(imagePath string) (domain.Image, error)
	// OpenFile gets image form local storage returning readable File struct
	OpenFile(imagePath string) (usecases.File, error)
	// InitImageListScanner initialize scanner to read image list from file
	InitImageListScanner(f usecases.File)
	// GetLocalImageListElement gets tuple element from image List, element format must be
	// [date][space][imagepath]
	GetLocalImageListElement() string
	// NextImageListElement returns true if there is more elements in Image List, otherwise returns false
	NextImageListElement() bool
	// ErrorScanningImageList returns error if the process of get element from image list failed
	ErrorScanningImageList() error
}

// CLIYamsLogger logs CLI yams events
type CLIYamsLogger interface {
	LogImage(int, usecases.YamsObject)
	LogErrorGettingImagesList(listPath string, err error)
	LogErrorCleaningMarks(imgName string, err error)
	LogErrorRemoteDelete(imgName string, err error)
	LogErrorResetingErrorCounter(imgName string, err error)
	LogErrorIncreasingErrorCounter(imgName string, err error)
	LogErrorGettingRemoteChecksum(imgName string, err error)
}

// retryPreviousFailedUploads gets images from errorControlRepository and try
// to upload those images to yams one more time. If fails increase the counter of errors
// in repo. Repository only returns images with less than a specific number of errors.
func (cli *CLIYams) retryPreviousFailedUploads(threads, maxErrorTolerance int) {
	maxConcurrency := cli.yamsService.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup
	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWRetry)
	}
	nPages := cli.errorControl.GetErrorsPagesQty(maxErrorTolerance)
	for pagination := 1; pagination <= nPages; pagination++ {
		result, err := cli.errorControl.GetPreviousErrors(pagination, maxErrorTolerance)
		if err != nil {
			continue
		}
		for _, imagePath := range result {
			image, err := cli.localImage.GetLocalImage(imagePath)
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
	maxConcurrency := cli.yamsService.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}

	cli.retryPreviousFailedUploads(threads, maxErrorQty)

	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}

	// Get the data file with list of images to upload
	file, e := cli.localImage.OpenFile(imagesDumpYamsPath)
	if e != nil {
		cli.logger.LogErrorGettingImagesList(imagesDumpYamsPath, e)
		return e
	}
	defer file.Close() // nolint

	latestSynchronizedImageDate := cli.lastSync.GetLastSynchronizationMark()
	var imagePath, imageDateStr string

	cli.localImage.InitImageListScanner(file)
	// for each element read from file
	for cli.localImage.NextImageListElement() {
		tuple := strings.Split(cli.localImage.GetLocalImageListElement(), " ")
		if !validateTuple(tuple, latestSynchronizedImageDate, cli.dateLayout) {
			continue
		}
		imageDateStr, imagePath = tuple[0], tuple[1]
		image, err := cli.localImage.GetLocalImage(imagePath)
		if err != nil {
			continue
		}
		jobs <- image
	}

	close(jobs)
	waitGroup.Wait()

	// If scanner stops because error
	if e := cli.localImage.ErrorScanningImageList(); e != nil {
		return fmt.Errorf("Error reading data from file: %+v", e)
	}
	err := cli.lastSync.SetLastSynchronizationMark(imageDateStr)
	if err != nil {
		return fmt.Errorf("Error setting synchronization mark %+v", err)
	}

	return nil
}

// validateTuple validates a given tuple string is format []string{dateStr,path}
// and the date after or before of a given date
func validateTuple(tuple []string, date time.Time, dateLayout string) bool {
	if len(tuple) != 2 {
		return false
	}
	imageDateStr := tuple[0]
	imageDate, err := time.Parse(dateLayout, imageDateStr)
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
	list, err := cli.yamsService.List()
	for i, img := range list {
		cli.logger.LogImage(i+1, img)
	}
	return err
}

// Delete deletes an object in yams repository
func (cli *CLIYams) Delete(imageName string) error {
	return cli.yamsService.RemoteDelete(imageName, domain.YAMSForceRemoval)
}

// DeleteAll deletes every imagen in yams repository and redis using concurency
func (cli *CLIYams) DeleteAll(threads int) error {
	images, err := cli.yamsService.List()
	if err != nil {
		return err
	}

	jobs := make(chan string)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
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
	defer wg.Done()
	for image := range jobs {
		err := cli.yamsService.Send(image)
		cli.sendErrorControl(image, previousUploadFailed, err)
	}
}

// sendErrorControl takes action depending of error type retuned by send method
func (cli *CLIYams) sendErrorControl(image domain.Image, previousUploadFailed int, err error) {
	imageName := image.Metadata.ImageName
	localImageChecksum := image.Metadata.Checksum
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)
	switch err {
	case nil:
		fallthrough
	case yamsErrNil:
		if previousUploadFailed == domain.SWRetry {
			if e := cli.errorControl.CleanErrorMarks(imageName); e != nil {
				cli.logger.LogErrorCleaningMarks(imageName, e)
			}
		}
		return
	case usecases.ErrYamsDuplicate:
		remoteImgChecksum, e := cli.yamsService.GetRemoteChecksum(imageName)
		if e != yamsErrNil {
			cli.logger.LogErrorGettingRemoteChecksum(imageName, e)
			// recursive increase error counter
			cli.sendErrorControl(image, previousUploadFailed, fmt.Errorf("error getting checksum"))
			return
		}
		if remoteImgChecksum != localImageChecksum {
			if e := cli.yamsService.RemoteDelete(imageName, domain.YAMSForceRemoval); e != yamsErrNil {
				cli.logger.LogErrorRemoteDelete(imageName, e)
				// recursive increase error counter
				cli.sendErrorControl(image, previousUploadFailed, fmt.Errorf("error deleting remote image"))
				return
			}
			// mark to upload in the next sync process (because yams cache)
			if e := cli.errorControl.SetErrorCounter(imageName, 0); e != nil {
				cli.logger.LogErrorResetingErrorCounter(imageName, e)
			}
		} else {
			// recursive clean up marks with nil error
			cli.sendErrorControl(image, previousUploadFailed, nil)
		}
	default: // any other kind of error increase error counter
		if e := cli.errorControl.IncreaseErrorCounter(imageName); e != nil {
			cli.logger.LogErrorIncreasingErrorCounter(imageName, e)
		}
	}
}

// deleteWorker deletes every image to yams repository
func (cli *CLIYams) deleteWorker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	for j := range jobs {
		if e := cli.yamsService.RemoteDelete(j, domain.YAMSForceRemoval); e != nil {
			cli.logger.LogErrorRemoteDelete(j, e)
		}
	}
	defer wg.Done()
}

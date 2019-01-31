package interfaces

import (
	"strings"
	"sync"
	"time"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// CLIYams is a yams client that executes operation on yams repository
type CLIYams struct {
	imageService         ImageService
	errorControl         ErrorControl
	lastSync             LastSync
	localImage           LocalImage
	logger               CLIYamsLogger
	dateLayout           string
	lastSyncDate         chan time.Time
	inProgressTimestamps chan []time.Time
	stats                Stats
	quit                 chan bool
	isSync               bool
	isDelete             bool
}

// NewCLIYams creates a new instance of CLIYams
func NewCLIYams(imageService ImageService, errorControl ErrorControl, lastSync LastSync,
	localImage LocalImage, logger CLIYamsLogger, defaultLastSyncDate time.Time, stats Stats, dateLayout string) *CLIYams {

	lastSyncDate := make(chan time.Time, 1)
	lastSyncDate <- defaultLastSyncDate
	quit := make(chan bool, 1)
	quit <- false
	inProgressTimestamps := make(chan []time.Time, 1)
	inProgressTimestamps <- []time.Time{}

	return &CLIYams{
		imageService:         imageService,
		errorControl:         errorControl,
		lastSync:             lastSync,
		localImage:           localImage,
		logger:               logger,
		dateLayout:           dateLayout,
		lastSyncDate:         lastSyncDate,
		inProgressTimestamps: inProgressTimestamps,
		quit:                 quit,
		stats:                stats,
	}
}

// ImageService allows operations between local repository & remote yams repository
type ImageService interface {
	// GetRemoteChecksum gets the checksum of image in YAMS
	GetRemoteChecksum(imageName string) (string, *usecases.YamsRepositoryError)
	// Send sends images from local storage to yams bucket
	Send(image domain.Image) (checksum string, err *usecases.YamsRepositoryError)
	// List gets list of available images in yams bucket
	List(oldContinuationToken string, step int) (images []usecases.YamsObject, newContinuationToken string, err *usecases.YamsRepositoryError)
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
	SetLastSynchronizationMark(date time.Time) error
	// Reset resets every synchronization mark
	Reset() error
	// Get gets list of synchronization marks
	Get() ([]string, error)
}

// LocalImage allows operations over local storage
type LocalImage interface {
	// GetLocalImage gets image form local storage parsed as domain.Image
	GetLocalImage(imagePath string) (domain.Image, error)
	// OpenFile gets image form local storage returning readable File struct
	OpenFile(imagePath string) (usecases.File, error)
	// InitImageListScanner initialize scanner to read image list from file
	InitImageListScanner(f usecases.File) Scanner
}

// Scanner allows operations to read file line by line
type Scanner interface {
	Text() string
	Scan() bool
	Err() error
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
	LogErrorSettingSyncMark(mark time.Time, err error)
	LogRetryPreviousFailedUploads()
	LogReadingNewImages()
	LogUploadingNewImages()
	LogStats(timer int, stats *Stats)
	LogMarksList(list []string)
}

// retryPreviousFailedUploads gets images from errorControlRepository and try
// to upload those images to yams one more time. If fails increase the counter of errors
// in repo. Repository only returns images with less than a specific number of errors.
func (cli *CLIYams) retryPreviousFailedUploads(threads, maxErrorTolerance int, latestSynchronizedImageDate time.Time) {
	maxConcurrency := cli.imageService.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}
	// Prepare the jobs using concurrency
	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup
	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
		go cli.retrySendWorker(w, jobs, &waitGroup)
	}
	// Get how many pages of failed uploads are in DB
	nPages := cli.errorControl.GetErrorsPagesQty(maxErrorTolerance)
	for pagination := 1; pagination <= nPages; pagination++ {
		// Get a list of failed uploads
		result, err := cli.errorControl.GetPreviousErrors(pagination, maxErrorTolerance)
		if err != nil {
			continue
		}
		// For each image in the list of failed uplaods
		for _, imagePath := range result {
			cli.stats.Processed <- inc(<-cli.stats.Processed)
			cli.stats.exposer.IncrementCounter(domain.ProcessedImages)
			image, err := cli.localImage.GetLocalImage(imagePath)
			if err != nil {
				cli.stats.NotFound <- inc(<-cli.stats.NotFound)
				cli.stats.exposer.IncrementCounter(domain.NotFoundImages)
				continue
			}

			imageDate := removeTimezoneDiff(image.Metadata.ModTime)
			// If the failed image is after of the last sync mark, remove the error mark
			// because this image will be uploaded when the sync process starts
			if imageDate.Equal(latestSynchronizedImageDate) ||
				imageDate.After(latestSynchronizedImageDate) {
				cli.stats.Recovered <- inc(<-cli.stats.Recovered)
				cli.stats.exposer.IncrementCounter(domain.RecoveredImages)
				if e := cli.errorControl.CleanErrorMarks(image.Metadata.ImageName); e != nil {
					cli.logger.LogErrorCleaningMarks(image.Metadata.ImageName, e)
				}
				continue
			}
			// Concurrent upload to imageService
			jobs <- image
		}
	}

	close(jobs)
	waitGroup.Wait()
}

// Sync synchronizes images between local repository and image service repository
// using go concurrency
func (cli *CLIYams) Sync(threads, syncLimit, maxErrorTolerance int, imagesDumpYamsPath string) error {
	cli.isSync = true
	maxConcurrency := cli.imageService.GetMaxConcurrency()
	if threads > maxConcurrency {
		threads = maxConcurrency
	}
	cli.showStats()
	cli.logger.LogRetryPreviousFailedUploads()

	latestSynchronizedImageDate := cli.lastSync.GetLastSynchronizationMark()
	<-cli.lastSyncDate
	cli.lastSyncDate <- latestSynchronizedImageDate

	// Retry failed uploads in previous synchronization process
	cli.retryPreviousFailedUploads(threads, maxErrorTolerance, latestSynchronizedImageDate)

	// prepare to upload using concurrent workers
	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup
	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}

	cli.logger.LogReadingNewImages()

	// Get the data file with list of images to upload
	file, e := cli.localImage.OpenFile(imagesDumpYamsPath)
	if e != nil {
		cli.logger.LogErrorGettingImagesList(imagesDumpYamsPath, e)
		return e
	}
	defer file.Close() // nolint

	cli.logger.LogUploadingNewImages()

	scanner := cli.localImage.InitImageListScanner(file)
	// for each element read from file
	for scanner.Scan() {
		cli.stats.Processed <- inc(<-cli.stats.Processed)
		cli.stats.exposer.IncrementCounter(domain.ProcessedImages)
		sentImages := <-cli.stats.Sent
		cli.stats.Sent <- sentImages
		if sentImages > syncLimit && syncLimit > 0 {
			break
		}
		tuple := strings.Split(scanner.Text(), " ")
		if !validateTuple(tuple, latestSynchronizedImageDate, cli.dateLayout) {
			cli.stats.Skipped <- inc(<-cli.stats.Skipped)
			cli.stats.exposer.IncrementCounter(domain.SkippedImages)
			continue
		}
		_, imagePath := tuple[0], tuple[1]
		image, err := cli.localImage.GetLocalImage(imagePath)
		if err != nil {
			cli.stats.NotFound <- inc(<-cli.stats.NotFound)
			cli.stats.exposer.IncrementCounter(domain.NotFoundImages)
			continue
		}
		jobs <- image
	}

	close(jobs)
	waitGroup.Wait()

	// If scanner stopped because error
	if e := scanner.Err(); e != nil {
		return e
	}

	// When the process is done, retry failed uploads using the new latestSynchronizedImageDate
	latestSynchronizedImageDate = <-cli.lastSyncDate
	cli.lastSyncDate <- latestSynchronizedImageDate
	cli.retryPreviousFailedUploads(threads, maxErrorTolerance, latestSynchronizedImageDate)
	return nil
}

// validateTuple validates a given tuple string is format []string{dateStr,path}
// and the date after or equal of a given date
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
func (cli *CLIYams) List(limit int) (err error) {
	counter := 0
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)
	var continuationToken, backupToken string
	var list []usecases.YamsObject
	// While images Service has images, list all of them,
	for {
		list, continuationToken, err = cli.imageService.List(continuationToken, 0)
		if err != yamsErrNil {
			if err == usecases.ErrYamsInternal {
				continuationToken = backupToken
			}
			continue
		}
		for _, image := range list {
			cli.logger.LogImage(counter+1, image)
			counter++
			if counter >= limit && limit > 0 {
				return nil
			}
		}
		// Empty continuationToken means no more pagination
		if continuationToken == "" {
			return nil
		}
		backupToken = continuationToken
	}
}

// Delete deletes an object in yams repository
func (cli *CLIYams) Delete(imageName string) error {
	return cli.imageService.RemoteDelete(imageName, domain.YAMSForceRemoval)
}

// DeleteAll deletes every imagen in yams repository and redis using concurency
func (cli *CLIYams) DeleteAll(threads, limit int) (err error) {
	cli.isDelete = true
	cli.showStats()
	jobs := make(chan domain.Image)
	var waitGroup sync.WaitGroup

	for w := 0; w < threads; w++ {
		waitGroup.Add(1)
		go cli.deleteWorker(w, jobs, &waitGroup)
	}

	// prepare to upload using concurrent workers
	latestSynchronizedImageDate := cli.lastSync.GetLastSynchronizationMark()
	<-cli.lastSyncDate
	cli.lastSyncDate <- latestSynchronizedImageDate

	yamsErrNil := (*usecases.YamsRepositoryError)(nil)
	var list []usecases.YamsObject
	var continuationToken string
	var backupToken string
	var counter int

	// While images Service has images, delete all of them
	for {
		list, continuationToken, err = cli.imageService.List(continuationToken, 0)
		if err != yamsErrNil {
			if err == usecases.ErrYamsInternal {
				continuationToken = backupToken
			}
			continue
		}
		for _, yamsObject := range list {
			cli.stats.Processed <- inc(<-cli.stats.Processed)
			cli.stats.exposer.IncrementCounter(domain.ProcessedImages)
			image, e := cli.localImage.GetLocalImage(yamsObject.ID)
			if e != nil {
				image.Metadata.ImageName = yamsObject.ID
				image.Metadata.ModTime = time.Now()
			}
			image.Metadata.ModTime = removeTimezoneDiff(image.Metadata.ModTime)
			jobs <- image
			counter++
			if counter >= limit && limit > 0 {
				break
			}
		}
		// Empty continuationToken means no more pagination
		if continuationToken == "" || (counter >= limit && limit > 0) {
			break
		}
		backupToken = continuationToken
	}
	close(jobs)
	waitGroup.Wait()
	if err == yamsErrNil {
		err = nil
	}
	return err
}

// sendWorker sends every image to yams repository
func (cli *CLIYams) sendWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup, previousUploadFailed int) {
	defer wg.Done()
	yamsNilResponse := (*usecases.YamsRepositoryError)(nil)
	for image := range jobs {
		// get images in progress & add new in progress image
		inProgress := <-cli.inProgressTimestamps
		inProgress = append(inProgress, image.Metadata.ModTime)

		cli.inProgressTimestamps <- inProgress

		// send new image to Image Service
		remoteChecksum, err := cli.imageService.Send(image)
		cli.sendErrorControl(image, previousUploadFailed, remoteChecksum, err)

		// remove sent timestamp image of inProgress list
		inProgress = <-cli.inProgressTimestamps
		inProgress = removeElement(image.Metadata.ModTime, inProgress)
		cli.inProgressTimestamps <- inProgress

		// Update latest sync mark only if yams returns no error
		if err == yamsNilResponse || err == usecases.ErrYamsDuplicate || err == nil {
			date := <-cli.lastSyncDate
			if image.Metadata.ModTime.After(date) {
				date = image.Metadata.ModTime
			}
			cli.lastSyncDate <- date
		}

		// determine if the worker should finish
		if quit, ok := <-cli.quit; ok {
			cli.quit <- quit
			if quit {
				return
			}
		} else {
			return
		}
	}
}

// retrySendWorker retry to send failed uploads to yams repository
func (cli *CLIYams) retrySendWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup) {
	defer wg.Done()
	for image := range jobs {
		// Retry to upload image to Image Service
		remoteChecksum, err := cli.imageService.Send(image)
		cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, err)
		// determine if the worker should finish
		if quit, ok := <-cli.quit; ok {
			cli.quit <- quit
			if quit {
				return
			}
		} else {
			return
		}
	}
}

// sendErrorControl takes action depending of error type retuned by send method
func (cli *CLIYams) sendErrorControl(image domain.Image, previousUploadFailed int, remoteChecksum string, err error) {
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
			cli.stats.Recovered <- inc(<-cli.stats.Recovered)
			cli.stats.exposer.IncrementCounter(domain.RecoveredImages)
		}
		cli.stats.Sent <- inc(<-cli.stats.Sent)
		cli.stats.exposer.IncrementCounter(domain.SentImages)
		return
	case usecases.ErrYamsDuplicate:
		cli.stats.Duplicated <- inc(<-cli.stats.Duplicated)
		cli.stats.exposer.IncrementCounter(domain.DuplicatedImages)
		if remoteChecksum != localImageChecksum {
			if e := cli.imageService.RemoteDelete(imageName, domain.YAMSForceRemoval); e != yamsErrNil {
				cli.logger.LogErrorRemoteDelete(imageName, e)
				// recursive increase error counter
				cli.sendErrorControl(image, previousUploadFailed, remoteChecksum, e)
				return
			}
			// mark to upload in the next sync process (because yams cache)
			if e := cli.errorControl.SetErrorCounter(imageName, 0); e != nil {
				cli.logger.LogErrorResetingErrorCounter(imageName, e)
			}
		} else {
			// recursive clean up marks with nil error in case of presviousUploadFailed true
			cli.sendErrorControl(image, previousUploadFailed, remoteChecksum, nil)
		}
	default: // any other kind of error increase error counter
		cli.stats.Errors <- inc(<-cli.stats.Errors)
		cli.stats.exposer.IncrementCounter(domain.FailedUploads)
		if e := cli.errorControl.IncreaseErrorCounter(imageName); e != nil {
			cli.logger.LogErrorIncreasingErrorCounter(imageName, e)
		}
	}
}

// deleteWorker deletes every image to yams repository
func (cli *CLIYams) deleteWorker(id int, jobs <-chan domain.Image, wg *sync.WaitGroup) {
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)
	for image := range jobs {
		if e := cli.imageService.RemoteDelete(image.Metadata.ImageName, domain.YAMSForceRemoval); e != yamsErrNil {
			cli.logger.LogErrorRemoteDelete(image.Metadata.ImageName, e)
		} else {
			date := <-cli.lastSyncDate
			if image.Metadata.ModTime.Before(date) {
				date = image.Metadata.ModTime
			}
			cli.lastSyncDate <- date
		}
		quit := <-cli.quit
		cli.quit <- quit
		if quit {
			return
		}
	}
	defer wg.Done()
}

// Reset cleans the last synchronization date mark to return to the previous
// synchronization status
func (cli *CLIYams) Reset() (err error) {
	return cli.lastSync.Reset()
}

// GetMarks gets list of synchronization marks ordered by newer to older
func (cli *CLIYams) GetMarks() error {
	list, err := cli.lastSync.Get()
	if err != nil {
		return err
	}
	cli.logger.LogMarksList(list)
	return nil
}

// Close closes cliYams execution
func (cli *CLIYams) Close() (err error) {
	if cli.isSync || cli.isDelete {
		newMark := <-cli.lastSyncDate
		cli.lastSyncDate <- newMark
		oldMark := cli.lastSync.GetLastSynchronizationMark()
		var condition bool
		if cli.isSync {
			condition = newMark.After(oldMark)
		} else if cli.isDelete {
			condition = newMark.Before(oldMark)
		}
		if condition {
			inProgress := <-cli.inProgressTimestamps
			// Search if images in progress have an older date mark
			for _, timestamp := range inProgress {
				if timestamp.Before(newMark) {
					newMark = timestamp
				}
			}
			cli.inProgressTimestamps <- inProgress
			err = cli.lastSync.SetLastSynchronizationMark(newMark)
			if err != nil {
				cli.logger.LogErrorSettingSyncMark(newMark, err)
			}
		}
		quit := <-cli.quit
		cli.quit <- !quit
	}
	return
}

// showStats displays synchronization stats in screen while yams-dav-sync script is running
func (cli *CLIYams) showStats() {
	go func() {
		quit, ok := <-cli.quit
		if ok {
			cli.quit <- quit
		}
		timer := 0
		ticker := time.NewTicker(time.Second)
		for !quit {
			cli.logger.LogStats(timer, &cli.stats)
			<-ticker.C
			timer++
			quit, ok := <-cli.quit
			if ok {
				cli.quit <- quit
			}
		}
		cli.logger.LogStats(timer, &cli.stats)
		cli.stats.Close() // nolint
		close(cli.quit)
		close(cli.inProgressTimestamps)
		close(cli.lastSyncDate)
		ticker.Stop()
	}()
}

// removeTimezoneDiff removes difference between image timezone and synchronization mark
func removeTimezoneDiff(imageDate time.Time) time.Time {
	_, diff := imageDate.Zone()
	return imageDate.
		UTC().
		Add(time.Duration(float64(diff)) * time.Second).
		Truncate(time.Second)
}

// removeElement removes a given element from slice
func removeElement(element time.Time, slice []time.Time) []time.Time {
	for i := 0; i < len(slice); i++ {
		if slice[i].Equal(element) {
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}

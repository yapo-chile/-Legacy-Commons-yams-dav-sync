package interfaces

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type mockImageService struct {
	mock.Mock
}

func (m *mockImageService) ValidateChecksum(image domain.Image) bool {
	args := m.Called(image)
	return args.Bool(0)
}

func (m *mockImageService) Send(image domain.Image) (string, *usecases.YamsRepositoryError) {
	args := m.Called(image)
	return args.String(0), args.Get(1).(*usecases.YamsRepositoryError)
}

func (m *mockImageService) List(continuationToken string, limit int) ([]usecases.YamsObject, string, *usecases.YamsRepositoryError) {
	args := m.Called(continuationToken, limit)
	return args.Get(0).([]usecases.YamsObject), args.String(1), args.Get(2).(*usecases.YamsRepositoryError)
}

func (m *mockImageService) RemoteDelete(imageName string, force bool) *usecases.YamsRepositoryError {
	args := m.Called(imageName, force)
	return args.Get(0).(*usecases.YamsRepositoryError)
}

func (m *mockImageService) GetMaxConcurrency() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockImageService) GetRemoteChecksum(imgName string) (string, *usecases.YamsRepositoryError) {
	args := m.Called(imgName)
	return args.String(0), args.Get(1).(*usecases.YamsRepositoryError)
}

type mockErrorControl struct {
	mock.Mock
}

func (m *mockErrorControl) GetErrorsPagesQty(tolerance int) int {
	args := m.Called(tolerance)
	return args.Int(0)
}

func (m *mockErrorControl) GetPreviousErrors(pagination, tolerance int) ([]string, error) {
	args := m.Called(pagination, tolerance)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockErrorControl) CleanErrorMarks(imageName string) error {
	args := m.Called(imageName)
	return args.Error(0)
}

func (m *mockErrorControl) SetErrorCounter(imageName string, counter int) error {
	args := m.Called(imageName, counter)
	return args.Error(0)
}

func (m *mockErrorControl) IncreaseErrorCounter(imageName string) error {
	args := m.Called(imageName)
	return args.Error(0)
}

type mockLocalImage struct {
	mock.Mock
}

func (m *mockLocalImage) GetLocalImage(imagePath string) (domain.Image, error) {
	args := m.Called(imagePath)
	return args.Get(0).(domain.Image), args.Error(1)
}

func (m *mockLocalImage) OpenFile(imagePath string) (usecases.File, error) {
	args := m.Called(imagePath)
	return args.Get(0).(usecases.File), args.Error(1)
}

func (m *mockLocalImage) InitImageListScanner(f usecases.File) Scanner {
	args := m.Called(f)
	return args.Get(0).(Scanner)
}

type mockScanner struct {
	mock.Mock
}

func (m *mockScanner) Scan() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockScanner) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockScanner) Text() string {
	args := m.Called()
	return args.String(0)
}

type mockLastSync struct {
	mock.Mock
}

func (m *mockLastSync) GetLastSynchronizationMark() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *mockLastSync) SetLastSynchronizationMark(imageDate time.Time) error {
	args := m.Called(imageDate)
	return args.Error(0)
}

func (m *mockLastSync) Get() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockLastSync) Reset() error {
	args := m.Called()
	return args.Error(0)
}

type mockFile struct {
	mock.Mock
}

func (m *mockFile) Close() (err error) {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFile) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) LogImage(n int, obj usecases.YamsObject) {
	m.Called(n, obj)
}

func (m *mockLogger) LogErrorCleaningMarks(imgName string, err error) {
	m.Called(imgName, err)
}

func (m *mockLogger) LogErrorRemoteDelete(imgName string, err error) {
	m.Called(imgName, err)
}

func (m *mockLogger) LogErrorResetingErrorCounter(imgName string, err error) {
	m.Called(imgName, err)
}
func (m *mockLogger) LogErrorIncreasingErrorCounter(imgName string, err error) {
	m.Called(imgName, err)
}

func (m *mockLogger) LogErrorGettingRemoteChecksum(imgName string, err error) {
	m.Called(imgName, err)
}

func (m *mockLogger) LogErrorGettingImagesList(listPath string, err error) {
	m.Called(listPath, err)
}

func (m *mockLogger) LogErrorSettingSyncMark(mark time.Time, err error) {
	m.Called(mark, err)
}

func (m *mockLogger) LogRetryPreviousFailedUploads() {
	m.Called()
}

func (m *mockLogger) LogReadingNewImages() {
	m.Called()
}

func (m *mockLogger) LogUploadingNewImages() {
	m.Called()
}

func (m *mockLogger) LogStats(timer int, s *Stats) {
	m.Called(timer, s)
}

func (m *mockLogger) LogMarksList(list []string) {
	m.Called(list)
}

type mockMetricsExposer struct {
	mock.Mock
}

func (m *mockMetricsExposer) IncrementCounter(metric int) {
	m.Called(metric)
}

func (m *mockMetricsExposer) SetGauge(metric int, value float64) {
	m.Called(metric, value)
}

func (m *mockMetricsExposer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewSync(t *testing.T) {
	now := time.Now()
	date := make(chan time.Time, 1)
	date <- now
	expected := &CLIYams{}
	var metricsExposer MetricsExposer
	expected.lastSyncDate = date
	result := NewCLIYams(
		expected.imageService,
		expected.errorControl,
		expected.lastSync,
		expected.localImage,
		expected.logger,
		now,
		NewStats(metricsExposer),
		expected.dateLayout,
	)
	assert.ObjectsAreEqualValues(expected, result)
}

func TestSyncProcess(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mFile := &mockFile{}
	mScanner := &mockScanner{}
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}
	// images to send
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))

	imagesToRetrySend := []string{}

	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil)
	mLogger.On("LogRetryPreviousFailedUploads").Once()
	mLogger.On("LogReadingNewImages").Once()
	mLogger.On("LogUploadingNewImages").Once()

	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))

	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil).Once()
	mLocalImage.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).
		Return(mScanner).Once()

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	mLastSync.On("GetLastSynchronizationMark", mock.AnythingOfType("string")).Return(date)

	imageListElements := []string{
		"20190102T150405 1.jpg",
		"INVALID ELEMENT",
		"20190102T150405 2.jpg",
	}
	// Uploading New images
	for i, testCases := 0, len(imageListElements); i < testCases; i++ {
		switch i {
		case 0: // Happy case: Element read from image list and send to yams
			mScanner.On("Text").Return(imageListElements[i]).Once()
			mScanner.On("Scan").Return(true).Once()
			mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, nil).Once()
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", (*usecases.YamsRepositoryError)(nil))
		case 1: // Invalid tuple and skipped element
			mScanner.On("Text").Return(imageListElements[i]).Once()
			mScanner.On("Scan").Return(true).Once()
		case 2: // Image not found in local & skipped
			mScanner.On("Text").Return(imageListElements[i]).Once()
			mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, fmt.Errorf("error"))
			mScanner.On("Scan").Return(true).Once()
		}
	}
	mScanner.On("Err").Return(nil).Once()

	mScanner.On("Scan").Return(false).Once()

	mFile.On("Close").Return(nil)

	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)

	cli.Sync(3, 0, 1, "/")

	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mScanner.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestSyncProcessErrorScanning(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mScanner := &mockScanner{}
	mMetricsExposer := &mockMetricsExposer{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)
	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mLocalImage.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).
		Return(mScanner)

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")

	mLogger.On("LogRetryPreviousFailedUploads")
	mLogger.On("LogReadingNewImages")
	mLogger.On("LogUploadingNewImages")
	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))

	mLastSync.On("GetLastSynchronizationMark", mock.AnythingOfType("string")).Return(date)
	mScanner.On("Scan").Return(false).Once()
	mScanner.On("Err").Return(fmt.Errorf("err")).Once()

	mFile.On("Close").Return(nil)

	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)

	err := cli.Sync(3, 0, 1, "/")
	assert.Error(t, err)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mScanner.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestSyncOverTheLimit(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mScanner := &mockScanner{}
	mMetricsExposer := &mockMetricsExposer{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)
	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mLocalImage.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).
		Return(mScanner)

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")

	mLogger.On("LogRetryPreviousFailedUploads")
	mLogger.On("LogReadingNewImages")
	mLogger.On("LogUploadingNewImages")
	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))

	mLastSync.On("GetLastSynchronizationMark", mock.AnythingOfType("string")).Return(date)
	mScanner.On("Scan").Return(true).Once()
	mScanner.On("Err").Return(nil).Once()

	mFile.On("Close").Return(nil)

	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)
	<-cli.stats.Sent
	cli.stats.Sent <- 2

	err := cli.Sync(3, 1, 1, "/")

	assert.Nil(t, err)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mScanner.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestSyncErrorOpeningFile(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mMetricsExposer := &mockMetricsExposer{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	// images to send
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)

	mLogger.On("LogRetryPreviousFailedUploads")
	mLogger.On("LogReadingNewImages")
	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))

	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, fmt.Errorf("err"))
	mLogger.On("LogErrorGettingImagesList",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*errors.errorString"))

	layout := "20060102T150405"
	newDate, _ := time.Parse(layout, "20170102T150405")
	mLastSync.On("GetLastSynchronizationMark").Return(newDate.Add(time.Second - 1))
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)

	err := cli.Sync(3, 0, 1, "/")

	assert.Error(t, err)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestRetryPreviousFailedUploads(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))

	imagesToRetrySend := []string{"0.jpg", "1.jpg", "2.jpg"}
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil).Once()
	for i, testCases := 0, len(imagesToRetrySend); i < testCases; i++ {
		switch i {
		case 0: // Happy case: Everything OK
			mLocalImage.On("GetLocalImage", imagesToRetrySend[i]).
				Return(domain.Image{}, nil).Once()
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", yamsResponse).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
		case 1: // error getting local image
			err := fmt.Errorf("Error")
			mLocalImage.On("GetLocalImage", imagesToRetrySend[i]).
				Return(domain.Image{}, err).Once()
		case 2: // Image will be synchronized in this process and is not necessary to upload again
			err := fmt.Errorf("Error")
			image := domain.Image{
				Metadata: domain.ImageMetadata{
					ModTime: time.Now(), // new image
				},
			}
			mLogger.On("LogErrorCleaningMarks",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString"))
			mLocalImage.On("GetLocalImage", imagesToRetrySend[i]).
				Return(image, nil).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(err)

		}
	}
	layout := "20060102T150405"
	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)
	cli.retryPreviousFailedUploads(3, 1, newDate)

	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestRetryPreviousFailedUploadsErrorGettingErrors(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mMetricsExposer := &mockMetricsExposer{}

	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	err := fmt.Errorf("Error")
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, err).Once()

	layout := "20060102T150405"
	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		nil,
		nil,
		nil,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)
	cli.retryPreviousFailedUploads(3, 1, newDate.Add(time.Second-1))
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestErrorControl(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}

	layout := "20060102T150405"
	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(
		mImageService,
		mErrorControl,
		mLastSync,
		mLocalImage,
		mLogger,
		newDate,
		NewStats(mMetricsExposer),
		layout,
	)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))

	yamsErrNil := (*usecases.YamsRepositoryError)(nil)
	for i, testcases := 0, 7; i < testcases; i++ {
		image := domain.Image{}
		remoteChecksum := ""
		switch i {
		case 0: // Error nil, clean error marks ok
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).
				Return(nil).Once()
			cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, nil)

		case 1: // Error nil, clean error marks error
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).
				Return(fmt.Errorf("err")).Once()
			mLogger.On("LogErrorCleaningMarks", mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Once()
			cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, nil)

		case 2: // Error duplicated, different checksums
			image.Metadata.Checksum, remoteChecksum = "the same", "not the same"
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).
				Return(yamsErrNil).Once()
			mErrorControl.On("SetErrorCounter", mock.AnythingOfType("string"), 0).
				Return(nil).Once()
			cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, usecases.ErrYamsDuplicate)

		case 3: // Error duplicated, different checksums & error with remote delete
			image.Metadata.Checksum, remoteChecksum = "the same", "not the same"
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).
				Return(usecases.ErrYamsInternal).Once()
			mLogger.On("LogErrorRemoteDelete", mock.AnythingOfType("string"), usecases.ErrYamsInternal).
				Return().Once()
			mErrorControl.On("IncreaseErrorCounter", mock.AnythingOfType("string")).
				Return(nil).Once()
			cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, usecases.ErrYamsDuplicate)

		case 4: // Error duplicated, different checksums & error with SetErrorCounter()
			image.Metadata.Checksum, remoteChecksum = "the same", "not the same"
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).
				Return(yamsErrNil).Once()
			mErrorControl.On("SetErrorCounter", mock.AnythingOfType("string"), 0).
				Return(fmt.Errorf("error")).Once()
			mLogger.On("LogErrorResetingErrorCounter", mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Once()
			cli.sendErrorControl(image, domain.SWRetry, remoteChecksum, usecases.ErrYamsDuplicate)

		case 5: // Error duplicated, same checksums, skip because it was already uploaded
			image.Metadata.Checksum, remoteChecksum = "the same", "the same"
			cli.sendErrorControl(image, domain.SWUpload, remoteChecksum, usecases.ErrYamsDuplicate)
		case 6: // Error default, increase error counter error
			mErrorControl.On("IncreaseErrorCounter", mock.AnythingOfType("string")).
				Return(fmt.Errorf("error")).Once()
			mLogger.On("LogErrorIncreasingErrorCounter", mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Once()
			cli.sendErrorControl(image, domain.SWUpload, remoteChecksum, usecases.ErrYamsInternal)
		}
	}
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestValidateTuple(t *testing.T) {
	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	testCases := []struct {
		tuple    []string
		time     time.Time
		expected bool
	}{
		{
			tuple:    []string{"20060102T150405", "1.jpg"},
			time:     date,
			expected: false,
		},
		{
			tuple:    []string{"20180102T150405", "2.jpg"},
			time:     date,
			expected: true,
		},
		{
			tuple:    []string{"", "2.jpg"},
			time:     date,
			expected: false,
		},
		{
			tuple:    []string{},
			time:     date,
			expected: false,
		},
	}
	for _, test := range testCases {
		result := validateTuple(test.tuple, test.time, layout)
		assert.Equal(t, test.expected, result)
	}

}

func TestList(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mLogger := &mockLogger{}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "13"}}
	cli := CLIYams{imageService: mImageService, logger: mLogger}

	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).
		Return(yamsObjectResponse, "123", yamsErrResponse).Once()
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).
		Return(yamsObjectResponse, "123", usecases.ErrYamsInternal).Once()
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).
		Return(yamsObjectResponse, "", yamsErrResponse).Once()

	mLogger.On("LogImage",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("usecases.YamsObject"))
	err := cli.List(10)
	assert.NoError(t, err)
	mImageService.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestListOverTheLimit(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mLogger := &mockLogger{}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	yamsObjectResponse := []usecases.YamsObject{{ID: "1"}, {ID: "2"}} // two objects
	cli := CLIYams{imageService: mImageService, logger: mLogger}
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).
		Return(yamsObjectResponse, "", yamsErrResponse)
	mLogger.On("LogImage",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("usecases.YamsObject")).Once() // Log once
	err := cli.List(1) // request only one image
	assert.NoError(t, err)
	mImageService.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mImageService := &mockImageService{}
	cli := CLIYams{imageService: mImageService}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsErrResponse)
	err := cli.Delete("foto.jpg")
	assert.Nil(t, err)
	mImageService.AssertExpectations(t)
}

func TestDeleteAll(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mLogger := &mockLogger{}

	layout := "20060102T150405"
	newDate, _ := time.Parse(layout, "20170102T150405")
	cli := NewCLIYams(mImageService, nil, mLastSync, mLocalImage, mLogger, newDate, NewStats(mMetricsExposer), layout)
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "12"}, {ID: "12"}}
	yamsNilResponse := (*usecases.YamsRepositoryError)(nil)

	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))
	mLastSync.On("GetLastSynchronizationMark").Return(time.Now())
	// Get the list of images to delete
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return(yamsObjectResponse, "abc123", yamsNilResponse).Once()
	mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, nil).Once()
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsNilResponse).Once()
	mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, fmt.Errorf("err")).Once()
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(usecases.ErrYamsInternal).Once()
	// Get list page two but with error, keep the continuation token.
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return([]usecases.YamsObject{}, "", usecases.ErrYamsInternal).Once()

	// Get the list using continuation token and delete 4 images
	mImageService.On("List", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return(yamsObjectResponse, "abc123", yamsNilResponse).Once()
	mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, nil).Once()
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsNilResponse).Once()
	mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, nil).Once()
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsNilResponse).Once()

	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))
	mLogger.On("LogErrorRemoteDelete", mock.AnythingOfType("string"), mock.AnythingOfType("*usecases.YamsRepositoryError"))

	err := cli.DeleteAll(1, 4)
	assert.Nil(t, err)
	mImageService.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestCloseSync(t *testing.T) {
	t.Parallel()
	mLastSync := &mockLastSync{}
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}

	layout := "20060102T150405"
	cli := NewCLIYams(nil, nil, mLastSync, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	quit := <-cli.quit
	cli.quit <- !quit
	inProgress := <-cli.inProgressTimestamps
	inProgress = append(
		inProgress,
		time.Time{}, // Very old date
	)
	cli.inProgressTimestamps <- inProgress
	mLastSync.On("GetLastSynchronizationMark").Return(time.Now().Add(-2 * time.Hour))

	mLastSync.On("SetLastSynchronizationMark", mock.AnythingOfType("time.Time")).
		Return(fmt.Errorf("err"))
	mLogger.On("LogErrorSettingSyncMark",
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("*errors.errorString"))
	cli.isSync = true
	err := cli.Close()

	isClosed := func(ch <-chan time.Time) bool { // check if a chan time.Time is closed
		select {
		case <-ch:
			return true
		default:
		}
		return false
	}

	assert.True(t, isClosed(cli.lastSyncDate))
	assert.Error(t, err)
	mLogger.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestCloseDeleteAll(t *testing.T) {
	t.Parallel()
	mLastSync := &mockLastSync{}
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}

	layout := "20060102T150405"
	cli := NewCLIYams(nil, nil, mLastSync, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	quit := <-cli.quit
	cli.quit <- !quit

	mLastSync.On("GetLastSynchronizationMark").Return(time.Now().Add(2 * time.Hour))

	mLastSync.On("SetLastSynchronizationMark", mock.AnythingOfType("time.Time")).
		Return(fmt.Errorf("err"))
	mLogger.On("LogErrorSettingSyncMark",
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("*errors.errorString"))
	cli.isDelete = true
	err := cli.Close()
	assert.Error(t, err)
	mLogger.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestSendWorker(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}

	mLastSync := &mockLastSync{}
	var waitGroup sync.WaitGroup

	jobs := make(chan domain.Image)
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)

	sent := make(chan int, 1)
	sent <- 0
	mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", yamsErrNil)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))

	layout := "20060102T150405"

	cli := NewCLIYams(mImageService, nil, mLastSync, nil, nil, time.Now(), NewStats(mMetricsExposer), layout)

	for w := 0; w < 1; w++ {
		waitGroup.Add(1)
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}
	testImages := []string{"1.jpg", "2.jpg"}
	image := domain.Image{}
	for i, imageName := range testImages {
		image.Metadata.ImageName = imageName
		image.Metadata.ModTime = time.Now()
		jobs <- image
		quit := <-cli.quit
		if i > 0 { // in case of sys interrumption
			quit = true
		}
		cli.quit <- quit
	}
	mImageService.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestRetrySendWorker(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	var waitGroup sync.WaitGroup

	jobs := make(chan domain.Image)
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)

	sent := make(chan int, 1)
	sent <- 0
	mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", yamsErrNil)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))
	mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil)

	layout := "20060102T150405"

	cli := NewCLIYams(mImageService, mErrorControl, mLastSync, nil, nil, time.Now(), NewStats(mMetricsExposer), layout)
	<-cli.quit
	cli.quit <- true
	for w := 0; w < 1; w++ {
		waitGroup.Add(1)
		go cli.retrySendWorker(w, jobs, &waitGroup)
	}
	testImages := []string{"1.jpg"}
	image := domain.Image{}
	for _, imageName := range testImages {
		image.Metadata.ImageName = imageName
		image.Metadata.ModTime = time.Now()
		jobs <- image
	}
	close(jobs)
	waitGroup.Wait()
	mImageService.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
}

func TestRetryOnClosedChann(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	var waitGroup sync.WaitGroup

	jobs := make(chan domain.Image)
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)

	sent := make(chan int, 1)
	sent <- 0
	mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", yamsErrNil)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))
	mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil)
	layout := "20060102T150405"

	cli := NewCLIYams(mImageService, mErrorControl, mLastSync, nil, nil, time.Now(), NewStats(mMetricsExposer), layout)
	close(cli.quit)
	<-cli.quit
	for w := 0; w < 1; w++ {
		waitGroup.Add(1)
		go cli.retrySendWorker(w, jobs, &waitGroup)
	}
	testImages := []string{"1.jpg"}
	image := domain.Image{}
	for _, imageName := range testImages {
		image.Metadata.ImageName = imageName
		image.Metadata.ModTime = time.Now()
		jobs <- image
	}
	close(jobs)
	waitGroup.Wait()
	mImageService.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
}

func TestSendWorkerWithClosedChannel(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}

	var waitGroup sync.WaitGroup

	jobs := make(chan domain.Image)
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)

	sent := make(chan int, 1)
	sent <- 0
	mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return("", yamsErrNil)
	mMetricsExposer.On("IncrementCounter", mock.AnythingOfType("int"))
	layout := "20060102T150405"

	cli := NewCLIYams(mImageService, nil, nil, nil, nil, time.Now(), NewStats(mMetricsExposer), layout)

	for w := 0; w < 1; w++ {
		waitGroup.Add(1)
		go cli.sendWorker(w, jobs, &waitGroup, domain.SWUpload)
	}
	testImages := []string{"1.jpg", "2.jpg"}
	image := domain.Image{}
	for i, imageName := range testImages {
		image.Metadata.ImageName = imageName
		image.Metadata.ModTime = time.Now()
		jobs <- image
		quit := <-cli.quit
		if i > 0 { // in case of close the channel
			close(cli.quit)
		} else {
			cli.quit <- quit
		}
	}
	mImageService.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestDeleteWorker(t *testing.T) {
	t.Parallel()
	mImageService := &mockImageService{}
	mMetricsExposer := &mockMetricsExposer{}

	mLogger := &mockLogger{}

	var waitGroup sync.WaitGroup

	jobs := make(chan domain.Image)
	yamsErrNil := (*usecases.YamsRepositoryError)(nil)

	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsErrNil)

	layout := "20060102T150405"
	cli := NewCLIYams(mImageService, nil, nil, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	for w := 0; w < 1; w++ {
		waitGroup.Add(1)
		go cli.deleteWorker(w, jobs, &waitGroup)
	}

	testImages := []string{"1.jpg", "2.j:g"}
	for i, imageName := range testImages {
		jobs <- domain.Image{
			Metadata: domain.ImageMetadata{
				ImageName: imageName,
			},
		}
		quit := <-cli.quit
		if i > 0 { // in case of sys interrumption
			quit = true
		}
		cli.quit <- quit
	}
	mLogger.AssertExpectations(t)
	mImageService.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestShowStatsWithInterrumption(t *testing.T) {
	mLogger := &mockLogger{}
	mMetricsExposer := &mockMetricsExposer{}
	layout := "20060102T150405"
	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))
	cli := NewCLIYams(nil, nil, nil, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	cli.showStats()
	ticker := time.Tick(time.Second + time.Millisecond*500)
	<-ticker
	<-cli.quit
	cli.quit <- false
	<-ticker
	<-cli.quit
	cli.quit <- true
	<-ticker
	mLogger.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestShowStats(t *testing.T) {
	t.Parallel()
	mMetricsExposer := &mockMetricsExposer{}
	mLogger := &mockLogger{}
	layout := "20060102T150405"
	mLogger.On("LogStats", mock.AnythingOfType("int"), mock.AnythingOfType("*interfaces.Stats"))
	cli := NewCLIYams(nil, nil, nil, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	cli.showStats()
	ticker := time.Tick(time.Second + time.Millisecond*500)
	<-cli.quit
	cli.quit <- true
	<-ticker
	mLogger.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
}

func TestReset(t *testing.T) {
	mMetricsExposer := &mockMetricsExposer{}
	mLastSync := &mockLastSync{}
	mLogger := &mockLogger{}
	layout := "20060102T150405"
	mLastSync.On("Reset").Return(nil)
	cli := NewCLIYams(nil, nil, mLastSync, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	cli.Reset()
	mLogger.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
}

func TestGetMarks(t *testing.T) {
	mMetricsExposer := &mockMetricsExposer{}
	mLastSync := &mockLastSync{}
	mLogger := &mockLogger{}
	layout := "20060102T150405"
	mLastSync.On("Get").Return([]string{}, nil)
	mLogger.On("LogMarksList", mock.AnythingOfType("[]string"))
	cli := NewCLIYams(nil, nil, mLastSync, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	err := cli.GetMarks()
	assert.NoError(t, err)
	mLogger.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
}

func TestGetMarksError(t *testing.T) {
	mMetricsExposer := &mockMetricsExposer{}
	mLastSync := &mockLastSync{}
	mLogger := &mockLogger{}
	layout := "20060102T150405"
	mLastSync.On("Get").Return([]string{}, fmt.Errorf("err"))
	cli := NewCLIYams(nil, nil, mLastSync, nil, mLogger, time.Now(), NewStats(mMetricsExposer), layout)
	err := cli.GetMarks()
	assert.Error(t, err)
	mLogger.AssertExpectations(t)
	mMetricsExposer.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
}

func TestRemoveElement(t *testing.T) {
	element1, element2 := time.Now(), time.Time{}
	cases := []struct {
		slice    []time.Time
		expected []time.Time
		element  time.Time
	}{
		{
			slice:    []time.Time{element2, element1, element2, element2, element2, element1},
			expected: []time.Time{element1, element1},
			element:  element2,
		},
		{
			slice:    []time.Time{},
			expected: []time.Time{},
			element:  element1,
		},
		{
			slice:    []time.Time{element1, element1},
			expected: []time.Time{},
			element:  element1,
		},
	}
	for _, v := range cases {
		result := removeElement(v.element, v.slice)
		assert.Equal(t, v.expected, result)
	}
}

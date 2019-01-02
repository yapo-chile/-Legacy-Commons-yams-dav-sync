package interfaces

import (
	"fmt"
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

func (m *mockImageService) Send(image domain.Image) *usecases.YamsRepositoryError {
	args := m.Called(image)
	return args.Get(0).(*usecases.YamsRepositoryError)
}

func (m *mockImageService) List() ([]usecases.YamsObject, *usecases.YamsRepositoryError) {
	args := m.Called()
	return args.Get(0).([]usecases.YamsObject), args.Get(1).(*usecases.YamsRepositoryError)
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

func (m *mockLastSync) SetLastSynchronizationMark(imageDateStr string) error {
	args := m.Called(imageDateStr)
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

func TestNewSync(t *testing.T) {
	expected := &CLIYams{}
	result := NewCLIYams(expected.imageService,
		expected.errorControl,
		expected.lastSync,
		expected.localImage,
		expected.logger,
		expected.dateLayout)
	assert.Equal(t, expected, result)
}

func TestSyncProcess(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mFile := &mockFile{}
	mScanner := &mockScanner{}
	mLogger := &mockLogger{}
	// images to send
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	imagesToRetrySend := []string{"0.jpg", "1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg", "7.jpg"}

	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil)

	// retry previous failed uploads
	for i, testCases := 0, len(imagesToRetrySend); i < testCases; i++ {
		mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).
			Return(domain.Image{Metadata: domain.ImageMetadata{ImageName: imagesToRetrySend[i]}}, nil).Once()
		switch i {
		case 0: // Everything ok
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()

		case 1: // Sending duplicated img with equal checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
			//	mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsErrNil).Once()

		case 2: // Sending duplicated img with different checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsErrNil).Once()
			mErrorControl.On("SetErrorCounter", mock.AnythingOfType("string"), 0).Return(nil).Once()
		case 3: // Sending duplicated img with different checksum and error in remote delete
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsError).Once()
			mLogger.On("LogErrorRemoteDelete",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mErrorControl.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(nil).Once()
		case 4: // Sending duplicated img with different checksum and error in resetErrorCounter
			sendResponse := usecases.ErrYamsDuplicate
			err := fmt.Errorf("Error reseting counter")
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsErrNil).Once()
			mErrorControl.On("SetErrorCounter", mock.AnythingOfType("string"), 0).Return(err).Once()
			mLogger.On("LogErrorResetingErrorCounter",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		case 5: // Error getting remote checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsError).Once()
			mLogger.On("LogErrorGettingRemoteChecksum",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mErrorControl.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(nil).Once()
		case 6: // Error increasing error counter
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			err := fmt.Errorf("Error increasing counter")
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mImageService.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsError).Once()
			mLogger.On("LogErrorGettingRemoteChecksum",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mErrorControl.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(err).Once()
			mLogger.On("LogErrorIncreasingErrorCounter",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		case 7: // Error cleaning up marks
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			err := fmt.Errorf("Error cleaning error marks")
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(err).Once()
			mLogger.On("LogErrorCleaningMarks",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		}
	}

	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mLocalImage.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).Return(mScanner)

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
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return((*usecases.YamsRepositoryError)(nil)).Once()
		case 1: // Invalid tuple and skipped element
			mScanner.On("Text").Return(imageListElements[i]).Once()
			mScanner.On("Scan").Return(true).Once()
		case 2: // Image not found in local & skipped
			mScanner.On("Text").Return(imageListElements[i]).Once()
			mScanner.On("Scan").Return(true).Once()
			mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, fmt.Errorf("error")).Once()
		}
	}
	mScanner.On("Err").Return(nil).Once()

	mScanner.On("Scan").Return(false).Once()

	mFile.On("Close").Return(nil)

	mLastSync.On("SetLastSynchronizationMark", mock.AnythingOfType("string")).Return(nil)

	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
		lastSync:     mLastSync,
		localImage:   mLocalImage,
		logger:       mLogger,
		dateLayout:   layout}

	cli.Sync(3, 1, "/")

	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mScanner.AssertExpectations(t)
}

func TestSyncProcessErrorScanning(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mScanner := &mockScanner{}
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

	mLastSync.On("GetLastSynchronizationMark", mock.AnythingOfType("string")).Return(date)
	mScanner.On("Scan").Return(false).Once()
	mScanner.On("Err").Return(fmt.Errorf("err")).Once()

	mFile.On("Close").Return(nil)

	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
		lastSync:     mLastSync,
		localImage:   mLocalImage,
		logger:       mLogger,
		dateLayout:   layout}

	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mScanner.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestSyncProcessErrorSettingMark(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mFile := &mockFile{}
	mScanner := &mockScanner{}
	mLogger := &mockLogger{}
	// images to send
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)

	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mLocalImage.On("InitImageListScanner",
		mock.AnythingOfType("*interfaces.mockFile")).Return(mScanner)

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	mLastSync.On("GetLastSynchronizationMark", mock.AnythingOfType("string")).Return(date)

	mScanner.On("Scan").Return(false).Once()
	mScanner.On("Err").Return(nil).Once()

	mFile.On("Close").Return(nil)

	mLastSync.On("SetLastSynchronizationMark", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
		lastSync:     mLastSync,
		localImage:   mLocalImage,
		logger:       mLogger,
		dateLayout:   layout}
	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)

	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mScanner.AssertExpectations(t)
}

func TestSyncErrorOpeningFile(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	// images to send
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)

	mLocalImage.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, fmt.Errorf("err"))
	mLogger.On("LogErrorGettingImagesList",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*errors.errorString")).Return()
	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
		lastSync:     mLastSync,
		localImage:   mLocalImage,
		logger:       mLogger,
	}

	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestRetryPreviousFailedUploads(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}
	mLastSync := &mockLastSync{}
	mLocalImage := &mockLocalImage{}
	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	imagesToRetrySend := []string{"0.jpg", "1.jpg"}
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil).Once()
	for i, testCases := 0, len(imagesToRetrySend); i < testCases; i++ {
		switch i {
		case 0: // Happy case: Everything OK
			mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).
				Return(domain.Image{}, nil).Once()
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			mImageService.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mErrorControl.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
		case 1: // error getting local image
			err := fmt.Errorf("Error")
			mLocalImage.On("GetLocalImage", mock.AnythingOfType("string")).
				Return(domain.Image{}, err).Once()
		}
	}
	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
		lastSync:     mLastSync,
		localImage:   mLocalImage}
	cli.retryPreviousFailedUploads(3, 1)

	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
	mLocalImage.AssertExpectations(t)
	mLastSync.AssertExpectations(t)
}

func TestRetryPreviousFailedUploadsErrorGettingErrors(t *testing.T) {
	mImageService := &mockImageService{}
	mErrorControl := &mockErrorControl{}

	mImageService.On("GetMaxConcurrency").Return(1)
	mErrorControl.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	err := fmt.Errorf("Error")
	mErrorControl.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, err).Once()

	cli := CLIYams{imageService: mImageService,
		errorControl: mErrorControl,
	}
	cli.retryPreviousFailedUploads(3, 1)
	mImageService.AssertExpectations(t)
	mErrorControl.AssertExpectations(t)
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
	mImageService := &mockImageService{}
	mLogger := &mockLogger{}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}}
	cli := CLIYams{imageService: mImageService, logger: mLogger}
	mImageService.On("List").Return(yamsObjectResponse, yamsErrResponse)
	mLogger.On("LogImage",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("usecases.YamsObject")).Return()
	err := cli.List()
	assert.Nil(t, err)
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
	mImageService := &mockImageService{}
	mLogger := &mockLogger{}

	cli := CLIYams{imageService: mImageService, logger: mLogger}
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "12"}}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)

	mImageService.On("List").Return(yamsObjectResponse, yamsErrResponse)
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(yamsErrResponse).Once()
	mImageService.On("RemoteDelete", mock.AnythingOfType("string"), true).Return(usecases.ErrYamsInternal).Once()

	mLogger.On("LogErrorRemoteDelete", mock.AnythingOfType("string"), mock.AnythingOfType("*usecases.YamsRepositoryError"))
	err := cli.DeleteAll(100)
	assert.Nil(t, err)
	mImageService.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestDeleteAllListError(t *testing.T) {
	mImageService := &mockImageService{}

	cli := CLIYams{imageService: mImageService}
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "12"}}

	mImageService.On("List").Return(yamsObjectResponse, usecases.ErrYamsInternal)

	err := cli.DeleteAll(100)
	assert.Equal(t, usecases.ErrYamsInternal, err)
	mImageService.AssertExpectations(t)
}

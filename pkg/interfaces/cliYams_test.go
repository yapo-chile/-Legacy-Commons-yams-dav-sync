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

type mockInteractor struct {
	mock.Mock
}

func (m *mockInteractor) ValidateChecksum(image domain.Image) bool {
	args := m.Called(image)
	return args.Bool(0)
}

func (m *mockInteractor) Send(image domain.Image) *usecases.YamsRepositoryError {
	args := m.Called(image)
	return args.Get(0).(*usecases.YamsRepositoryError)
}

func (m *mockInteractor) List() ([]usecases.YamsObject, *usecases.YamsRepositoryError) {
	args := m.Called()
	return args.Get(0).([]usecases.YamsObject), args.Get(1).(*usecases.YamsRepositoryError)
}

func (m *mockInteractor) RemoteDelete(imageName string) *usecases.YamsRepositoryError {
	args := m.Called(imageName)
	return args.Get(0).(*usecases.YamsRepositoryError)
}

func (m *mockInteractor) GetMaxConcurrency() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockInteractor) GetRemoteChecksum(imgName string) (string, *usecases.YamsRepositoryError) {
	args := m.Called(imgName)
	return args.String(0), args.Get(1).(*usecases.YamsRepositoryError)
}

func (m *mockInteractor) GetErrorsPagesQty(tolerance int) int {
	args := m.Called(tolerance)
	return args.Int(0)
}

func (m *mockInteractor) GetPreviousErrors(pagination, tolerance int) ([]string, error) {
	args := m.Called(pagination, tolerance)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockInteractor) CleanErrorMarks(imageName string) error {
	args := m.Called(imageName)
	return args.Error(0)
}

func (m *mockInteractor) ResetErrorCounter(imageName string) error {
	args := m.Called(imageName)
	return args.Error(0)
}

func (m *mockInteractor) IncreaseErrorCounter(imageName string) error {
	args := m.Called(imageName)
	return args.Error(0)
}

func (m *mockInteractor) GetLocalImage(imagePath string) (domain.Image, error) {
	args := m.Called(imagePath)
	return args.Get(0).(domain.Image), args.Error(1)
}

func (m *mockInteractor) OpenFile(imagePath string) (usecases.File, error) {
	args := m.Called(imagePath)
	return args.Get(0).(usecases.File), args.Error(1)
}

func (m *mockInteractor) GetLastSynchornizationMark() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *mockInteractor) SetLastSynchornizationMark(imageDateStr string) error {
	args := m.Called(imageDateStr)
	return args.Error(0)
}

func (m *mockInteractor) InitImageListScanner(f usecases.File) {
	m.Called(f)
}

func (m *mockInteractor) NextImageListElement() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockInteractor) ErrorScanningImageList() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockInteractor) GetImageListElement() string {
	args := m.Called()
	return args.String(0)
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

func TestSyncProcess(t *testing.T) {
	mInteractor := &mockInteractor{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	// images to send
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	imagesToRetrySend := []string{"0.jpg", "1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg", "7.jpg"}

	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil)

	// retry previous failed uploads
	for i, testCases := 0, len(imagesToRetrySend); i < testCases; i++ {
		mInteractor.On("GetLocalImage", mock.AnythingOfType("string")).
			Return(domain.Image{Metadata: domain.ImageMetadata{ImageName: imagesToRetrySend[i]}}, nil).Once()
		switch i {
		case 0: // Everything ok
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()

		case 1: // Sending duplicated img with equal checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
			//	mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsErrNil).Once()

		case 2: // Sending duplicated img with different checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(yamsErrNil).Once()
			mInteractor.On("ResetErrorCounter", mock.AnythingOfType("string")).Return(nil).Once()
		case 3: // Sending duplicated img with different checksum and error in remote delete
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(yamsError).Once()
			mLogger.On("LogErrorRemoteDelete",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mInteractor.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(nil).Once()
		case 4: // Sending duplicated img with different checksum and error in resetErrorCounter
			sendResponse := usecases.ErrYamsDuplicate
			err := fmt.Errorf("Error reseting counter")
			yamsErrNil := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("algo", yamsErrNil).Once()
			mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(yamsErrNil).Once()
			mInteractor.On("ResetErrorCounter", mock.AnythingOfType("string")).Return(err).Once()
			mLogger.On("LogErrorResetingErrorCounter",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		case 5: // Error getting remote checksum
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsError).Once()
			mLogger.On("LogErrorGettingRemoteChecksum",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mInteractor.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(nil).Once()
		case 6: // Error increasing error counter
			sendResponse := usecases.ErrYamsDuplicate
			yamsError := usecases.ErrYamsInternal
			err := fmt.Errorf("Error increasing counter")
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(sendResponse).Once()
			mInteractor.On("GetRemoteChecksum", mock.AnythingOfType("string")).Return("", yamsError).Once()
			mLogger.On("LogErrorGettingRemoteChecksum",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*usecases.YamsRepositoryError")).Return().Once()
			mInteractor.On("IncreaseErrorCounter", mock.AnythingOfType("string")).Return(err).Once()
			mLogger.On("LogErrorIncreasingErrorCounter",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		case 7: // Error cleaning up marks
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			err := fmt.Errorf("Error cleaning error marks")
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(err).Once()
			mLogger.On("LogErrorCleaningMarks",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*errors.errorString")).Return().Once()
		}
	}

	mInteractor.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mInteractor.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).Return()

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	mInteractor.On("GetLastSynchornizationMark", mock.AnythingOfType("string")).Return(date)

	imageListElements := []string{
		"20190102T150405 1.jpg",
		"INVALID ELEMENT",
		"20190102T150405 2.jpg",
	}
	// Uploading New images
	for i, testCases := 0, len(imageListElements); i < testCases; i++ {
		switch i {
		case 0: // Happy case: Element read from image list and send to yams
			mInteractor.On("GetImageListElement").Return(imageListElements[i]).Once()
			mInteractor.On("NextImageListElement").Return(true).Once()
			mInteractor.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, nil).Once()
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return((*usecases.YamsRepositoryError)(nil)).Once()
		case 1: // Invalid tuple and skipped element
			mInteractor.On("GetImageListElement").Return(imageListElements[i]).Once()
			mInteractor.On("NextImageListElement").Return(true).Once()
		case 2: // Image not found in local & skipped
			mInteractor.On("GetImageListElement").Return(imageListElements[i]).Once()
			mInteractor.On("NextImageListElement").Return(true).Once()
			mInteractor.On("GetLocalImage", mock.AnythingOfType("string")).Return(domain.Image{}, fmt.Errorf("error")).Once()
		}
	}
	mInteractor.On("ErrorScanningImageList").Return(nil).Once()

	mInteractor.On("NextImageListElement").Return(false).Once()

	mFile.On("Close").Return(nil)

	mInteractor.On("SetLastSynchornizationMark", mock.AnythingOfType("string")).Return(nil)

	cli := CLIYams{Interactor: mInteractor, Logger: mLogger, DateLayout: layout}
	cli.Sync(3, 1, "/")
	mInteractor.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestSyncProcessErrorScanning(t *testing.T) {
	mInteractor := &mockInteractor{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)
	mInteractor.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mInteractor.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).
		Return()

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")

	mInteractor.On("GetLastSynchornizationMark", mock.AnythingOfType("string")).Return(date)
	mInteractor.On("NextImageListElement").Return(false).Once()
	mInteractor.On("ErrorScanningImageList").Return(fmt.Errorf("err")).Once()

	mFile.On("Close").Return(nil)

	cli := CLIYams{Interactor: mInteractor, Logger: mLogger, DateLayout: layout}
	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)
	mInteractor.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestSyncProcessErrorSettingMark(t *testing.T) {
	mInteractor := &mockInteractor{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	// images to send
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)

	mInteractor.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, nil)
	mInteractor.On("InitImageListScanner", mock.AnythingOfType("*interfaces.mockFile")).Return()

	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	mInteractor.On("GetLastSynchornizationMark", mock.AnythingOfType("string")).Return(date)

	mInteractor.On("NextImageListElement").Return(false).Once()
	mInteractor.On("ErrorScanningImageList").Return(nil).Once()

	mFile.On("Close").Return(nil)

	mInteractor.On("SetLastSynchornizationMark", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	cli := CLIYams{Interactor: mInteractor, Logger: mLogger, DateLayout: layout}
	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)
	mInteractor.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestSyncErrorOpeningFile(t *testing.T) {
	mInteractor := &mockInteractor{}
	mFile := &mockFile{}
	mLogger := &mockLogger{}
	// images to send
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)

	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, nil)

	mInteractor.On("OpenFile", mock.AnythingOfType("string")).Return(mFile, fmt.Errorf("err"))

	cli := CLIYams{Interactor: mInteractor, Logger: mLogger}
	err := cli.Sync(3, 1, "/")
	assert.Error(t, err)
	mInteractor.AssertExpectations(t)
	mLogger.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestRetryPreviousFailedUploads(t *testing.T) {
	mInteractor := &mockInteractor{}
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	imagesToRetrySend := []string{"0.jpg", "1.jpg"}
	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(imagesToRetrySend, nil).Once()
	for i, testCases := 0, len(imagesToRetrySend); i < testCases; i++ {
		switch i {
		case 0: // Happy case: Everything OK
			mInteractor.On("GetLocalImage", mock.AnythingOfType("string")).
				Return(domain.Image{}, nil).Once()
			yamsResponse := (*usecases.YamsRepositoryError)(nil)
			mInteractor.On("Send", mock.AnythingOfType("domain.Image")).Return(yamsResponse).Once()
			mInteractor.On("CleanErrorMarks", mock.AnythingOfType("string")).Return(nil).Once()
		case 1: // error getting local image
			err := fmt.Errorf("Error")
			mInteractor.On("GetLocalImage", mock.AnythingOfType("string")).
				Return(domain.Image{}, err).Once()
		}
	}
	cli := CLIYams{Interactor: mInteractor}
	cli.retryPreviousFailedUploads(3, 1)
	mInteractor.AssertExpectations(t)
}

func TestRetryPreviousFailedUploadsErrorGettingErrors(t *testing.T) {
	mInteractor := &mockInteractor{}
	mInteractor.On("GetMaxConcurrency").Return(1)
	mInteractor.On("GetErrorsPagesQty", mock.AnythingOfType("int")).Return(1)
	err := fmt.Errorf("Error")
	mInteractor.On("GetPreviousErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return([]string{}, err).Once()

	cli := CLIYams{Interactor: mInteractor}
	cli.retryPreviousFailedUploads(3, 1)
	mInteractor.AssertExpectations(t)
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
	mInteractor := &mockInteractor{}
	mLogger := &mockLogger{}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}}
	cli := CLIYams{Interactor: mInteractor, Logger: mLogger}
	mInteractor.On("List").Return(yamsObjectResponse, yamsErrResponse)
	mLogger.On("LogImage",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("usecases.YamsObject")).Return()
	err := cli.List()
	assert.Nil(t, err)
	mInteractor.AssertExpectations(t)
	mLogger.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mInteractor := &mockInteractor{}
	cli := CLIYams{Interactor: mInteractor}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)
	mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(yamsErrResponse)
	err := cli.Delete("foto.jpg")
	assert.Nil(t, err)
}

func TestDeleteAll(t *testing.T) {
	mInteractor := &mockInteractor{}
	mLogger := &mockLogger{}

	cli := CLIYams{Interactor: mInteractor, Logger: mLogger}
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "12"}}
	yamsErrResponse := (*usecases.YamsRepositoryError)(nil)

	mInteractor.On("List").Return(yamsObjectResponse, yamsErrResponse)
	mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(yamsErrResponse).Once()
	mInteractor.On("RemoteDelete", mock.AnythingOfType("string")).Return(usecases.ErrYamsInternal).Once()

	mLogger.On("LogErrorRemoteDelete", mock.AnythingOfType("string"), mock.AnythingOfType("*usecases.YamsRepositoryError"))
	err := cli.DeleteAll(100)
	assert.Nil(t, err)
}

func TestDeleteAllListError(t *testing.T) {
	mInteractor := &mockInteractor{}

	cli := CLIYams{Interactor: mInteractor}
	yamsObjectResponse := []usecases.YamsObject{{ID: "12"}, {ID: "12"}}

	mInteractor.On("List").Return(yamsObjectResponse, usecases.ErrYamsInternal)

	err := cli.DeleteAll(100)
	assert.Equal(t, usecases.ErrYamsInternal, err)
}

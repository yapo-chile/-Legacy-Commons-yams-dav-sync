package usecases

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

type mockYamsRepo struct {
	mock.Mock
}

func (m *mockYamsRepo) GetMaxConcurrentConns() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockYamsRepo) GetImages() ([]YamsObject, *YamsRepositoryError) {
	args := m.Called()
	return args.Get(0).([]YamsObject), args.Get(1).(*YamsRepositoryError)
}

func (m *mockYamsRepo) PutImage(img domain.Image) (err *YamsRepositoryError) {
	args := m.Called(img)
	return args.Get(0).(*YamsRepositoryError)
}

func (m *mockYamsRepo) HeadImage(imgName string) (hash string, err *YamsRepositoryError) {
	args := m.Called(imgName)
	return args.String(0), args.Get(1).(*YamsRepositoryError)
}

func (m *mockYamsRepo) DeleteImage(imgName string, inmediateRemoval bool) (err *YamsRepositoryError) {
	args := m.Called(imgName, inmediateRemoval)
	return args.Get(0).(*YamsRepositoryError)
}

func TestValidateChecksum(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}

	input := domain.Image{
		Metadata: domain.ImageMetadata{
			ImageName: "name",
			Checksum:  "123",
		},
	}
	mYamsRepo.On("HeadImage", "name").
		Return(input.Metadata.Checksum, &YamsRepositoryError{})
	result := sync.ValidateChecksum(input)
	assert.Equal(t, true, result)
	mYamsRepo.AssertExpectations(t)
}

func TestGetRemoteChecksum(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}

	input := domain.Image{
		Metadata: domain.ImageMetadata{
			ImageName: "name",
			Checksum:  "123",
		},
	}
	mYamsRepo.On("HeadImage", mock.AnythingOfType("string")).
		Return(input.Metadata.Checksum, &YamsRepositoryError{}, nil)
	result, err := sync.GetRemoteChecksum(input.Metadata.ImageName)
	assert.Equal(t, input.Metadata.Checksum, result)
	assert.Equal(t, &YamsRepositoryError{}, err)

	mYamsRepo.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}
	input := domain.Image{}
	mYamsRepo.On("PutImage", mock.AnythingOfType("domain.Image")).
		Return(&YamsRepositoryError{})
	result := sync.Send(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

func TestList(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}
	mYamsRepo.On("GetImages").Return([]YamsObject{}, &YamsRepositoryError{})
	result, err := sync.List()
	assert.Equal(t, []YamsObject{}, result)
	assert.Equal(t, &YamsRepositoryError{}, err)
	mYamsRepo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}
	input := "123.jpg"
	mYamsRepo.On(
		"DeleteImage",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("bool"),
	).Return(&YamsRepositoryError{})
	result := sync.RemoteDelete(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

func TestGetMaxConcurrency(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	sync := SyncInteractor{YamsRepo: &mYamsRepo}
	expected := 1
	mYamsRepo.On("GetMaxConcurrentConns").Return(expected)
	result := sync.GetMaxConcurrency()
	assert.Equal(t, expected, result)
	mYamsRepo.AssertExpectations(t)
}

type mockErrorControlRepo struct {
	mock.Mock
}

func (m *mockErrorControlRepo) GetPagesQty(tolerance int) int {
	args := m.Called(tolerance)
	return args.Int(0)
}

func (m *mockErrorControlRepo) GetSyncErrors(pagination, tolerance int) ([]string, error) {
	args := m.Called(pagination, tolerance)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockErrorControlRepo) DelSyncError(imgName string) error {
	args := m.Called(imgName)
	return args.Error(0)
}

func (m *mockErrorControlRepo) SetErrorCounter(imgName string, counter int) error {
	args := m.Called(imgName, counter)
	return args.Error(0)
}

func (m *mockErrorControlRepo) AddSyncError(imgName string) error {
	args := m.Called(imgName)
	return args.Error(0)
}

func TestGetPagesQty(t *testing.T) {
	mErrorControlRepo := mockErrorControlRepo{}
	sync := SyncInteractor{ErrorControlRepo: &mErrorControlRepo}
	expected := 1
	mErrorControlRepo.On("GetPagesQty", mock.AnythingOfType("int")).
		Return(expected)
	result := sync.GetErrorsPagesQty(1)
	assert.Equal(t, expected, result)
	mErrorControlRepo.AssertExpectations(t)
}

func TestGetPreviousErrors(t *testing.T) {
	mErrorControlRepo := mockErrorControlRepo{}
	sync := SyncInteractor{ErrorControlRepo: &mErrorControlRepo}
	expected := []string{"errorcito"}
	mErrorControlRepo.On(
		"GetSyncErrors",
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
	).Return(expected, nil)
	result, err := sync.GetPreviousErrors(1, 1)
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mErrorControlRepo.AssertExpectations(t)
}

func TestCleanErrorMarks(t *testing.T) {
	mErrorControlRepo := mockErrorControlRepo{}
	sync := SyncInteractor{ErrorControlRepo: &mErrorControlRepo}
	mErrorControlRepo.On(
		"DelSyncError",
		mock.AnythingOfType("string"),
	).Return(nil)
	err := sync.CleanErrorMarks("foto-sexy.jpg")
	assert.Nil(t, err)
	assert.Nil(t, err)
	mErrorControlRepo.AssertExpectations(t)
}

func TestResetErrorCounter(t *testing.T) {
	mErrorControlRepo := mockErrorControlRepo{}
	sync := SyncInteractor{ErrorControlRepo: &mErrorControlRepo}
	mErrorControlRepo.On(
		"SetErrorCounter",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("int"),
	).Return(nil)
	err := sync.ResetErrorCounter("foto-sexy.jpg")
	assert.Nil(t, err)
	assert.Nil(t, err)
	mErrorControlRepo.AssertExpectations(t)
}

func TestIncreaseErrorCounter(t *testing.T) {
	mErrorControlRepo := mockErrorControlRepo{}
	sync := SyncInteractor{ErrorControlRepo: &mErrorControlRepo}
	mErrorControlRepo.On(
		"AddSyncError",
		mock.AnythingOfType("string"),
	).Return(nil)
	err := sync.IncreaseErrorCounter("foto-sexy.jpg")
	assert.Nil(t, err)
	mErrorControlRepo.AssertExpectations(t)
}

type mockImageRepo struct {
	mock.Mock
}

func (m *mockImageRepo) GetImage(imgPath string) (domain.Image, error) {
	args := m.Called(imgPath)
	return args.Get(0).(domain.Image), args.Error(1)
}
func (m *mockImageRepo) OpenFile(imgPath string) (File, error) {
	args := m.Called(imgPath)
	return args.Get(0).(File), args.Error(1)
}

func (m *mockImageRepo) InitImageListScanner(f File) {
	m.Called(f)
}

func (m *mockImageRepo) NextImageListElement() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockImageRepo) ErrorScanningImageList() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockImageRepo) GetImageListElement() string {
	args := m.Called()
	return args.String(0)
}

func TestGetLocalImage(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	expected := domain.Image{}
	mImageRepo.On(
		"GetImage",
		mock.AnythingOfType("string"),
	).Return(expected, nil)
	result, err := sync.GetLocalImage("foto-sexy.jpg")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mImageRepo.AssertExpectations(t)
}

func TestOpenFileLocalImage(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	expected := &os.File{}
	mImageRepo.On(
		"OpenFile",
		mock.AnythingOfType("string"),
	).Return(expected, nil)
	result, err := sync.OpenFile("foto-sexy.jpg")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mImageRepo.AssertExpectations(t)
}

func TestInitImageListScanner(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	input := &os.File{}
	mImageRepo.On("InitImageListScanner", mock.AnythingOfType("*os.File")).Return()
	sync.InitImageListScanner(input)
	mImageRepo.AssertExpectations(t)
}

func TestGetImageListElement(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	expected := "fotillo.jpg"
	mImageRepo.On("GetImageListElement").Return(expected)
	result := sync.GetImageListElement()
	assert.Equal(t, expected, result)
	mImageRepo.AssertExpectations(t)
}
func TestNextImageListElement(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	expected := true
	mImageRepo.On("NextImageListElement").Return(expected)
	result := sync.NextImageListElement()
	assert.Equal(t, expected, result)
	mImageRepo.AssertExpectations(t)
}

func TestErrorScanningImageList(t *testing.T) {
	mImageRepo := mockImageRepo{}
	sync := SyncInteractor{ImageRepo: &mImageRepo}
	expected := fmt.Errorf("err")
	mImageRepo.On("ErrorScanningImageList").Return(expected)
	result := sync.ErrorScanningImageList()
	assert.Equal(t, expected, result)
	mImageRepo.AssertExpectations(t)
}

func TestNewSyncInteractor(t *testing.T) {
	expected := &SyncInteractor{}
	result := NewSyncInteractor(nil, nil, nil, nil)
	assert.Equal(t, expected, result)
}

type mockLastSyncRepo struct {
	mock.Mock
}

func (m *mockLastSyncRepo) GetLastSync() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *mockLastSyncRepo) SetLastSync(imgDate string) error {
	args := m.Called(imgDate)
	return args.Error(0)
}

func TestGetLastSynchornizationMark(t *testing.T) {
	mLastSyncRepo := mockLastSyncRepo{}
	sync := SyncInteractor{LastSyncRepo: &mLastSyncRepo}
	expected := time.Time{}
	mLastSyncRepo.On("GetLastSync").Return(expected, nil)
	result := sync.GetLastSynchornizationMark()
	assert.Equal(t, expected, result)
	mLastSyncRepo.AssertExpectations(t)
}

func TestSetLastSynchornizationMark(t *testing.T) {
	mLastSyncRepo := mockLastSyncRepo{}
	sync := SyncInteractor{LastSyncRepo: &mLastSyncRepo}
	mLastSyncRepo.On("SetLastSync", mock.AnythingOfType("string")).
		Return(nil)
	err := sync.SetLastSynchornizationMark("")
	assert.Nil(t, err)
	mLastSyncRepo.AssertExpectations(t)
}

func TestSGetImageListElement(t *testing.T) {
	mLastSyncRepo := mockLastSyncRepo{}
	sync := SyncInteractor{LastSyncRepo: &mLastSyncRepo}
	mLastSyncRepo.On("SetLastSync", mock.AnythingOfType("string")).
		Return(nil)
	err := sync.SetLastSynchornizationMark("")
	assert.Nil(t, err)
	mLastSyncRepo.AssertExpectations(t)
}

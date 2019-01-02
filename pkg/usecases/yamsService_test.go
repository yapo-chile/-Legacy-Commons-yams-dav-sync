package usecases

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

func TestError(t *testing.T) {
	repo := YamsRepositoryError{ErrorString: "err"}
	result := repo.Error()
	assert.Equal(t, result, "err")
}

type mockYamsRepo struct {
	mock.Mock
}

func (m *mockYamsRepo) GetMaxConcurrency() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockYamsRepo) GetLocalImages() ([]YamsObject, *YamsRepositoryError) {
	args := m.Called()
	return args.Get(0).([]YamsObject), args.Get(1).(*YamsRepositoryError)
}

func (m *mockYamsRepo) Send(img domain.Image) (err *YamsRepositoryError) {
	args := m.Called(img)
	return args.Get(0).(*YamsRepositoryError)
}

func (m *mockYamsRepo) GetRemoteChecksum(imgName string) (hash string, err *YamsRepositoryError) {
	args := m.Called(imgName)
	return args.String(0), args.Get(1).(*YamsRepositoryError)
}

func (m *mockYamsRepo) RemoteDelete(imgName string, inmediateRemoval bool) (err *YamsRepositoryError) {
	args := m.Called(imgName, inmediateRemoval)
	return args.Get(0).(*YamsRepositoryError)
}

func TestNewYamsService(t *testing.T) {
	expected := YamsService{}
	result := NewYamsService(nil)
	assert.Equal(t, expected, result)
}

func TestValidateChecksum(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}

	input := domain.Image{
		Metadata: domain.ImageMetadata{
			ImageName: "name",
			Checksum:  "123",
		},
	}
	mYamsRepo.On("GetRemoteChecksum", "name").
		Return(input.Metadata.Checksum, &YamsRepositoryError{})
	result := yamsService.ValidateChecksum(input)
	assert.Equal(t, true, result)
	mYamsRepo.AssertExpectations(t)
}

func TestGetRemoteChecksum(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}

	input := domain.Image{
		Metadata: domain.ImageMetadata{
			ImageName: "name",
			Checksum:  "123",
		},
	}
	mYamsRepo.On("GetRemoteChecksum", mock.AnythingOfType("string")).
		Return(input.Metadata.Checksum, &YamsRepositoryError{}, nil)
	result, err := yamsService.GetRemoteChecksum(input.Metadata.ImageName)
	assert.Equal(t, input.Metadata.Checksum, result)
	assert.Equal(t, &YamsRepositoryError{}, err)

	mYamsRepo.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}
	input := domain.Image{}
	mYamsRepo.On("Send", mock.AnythingOfType("domain.Image")).
		Return(&YamsRepositoryError{})
	result := yamsService.Send(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

func TestList(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}
	mYamsRepo.On("GetLocalImages").Return([]YamsObject{}, &YamsRepositoryError{})
	result, err := yamsService.List()
	assert.Equal(t, []YamsObject{}, result)
	assert.Equal(t, &YamsRepositoryError{}, err)
	mYamsRepo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}
	input := "123.jpg"
	mYamsRepo.On(
		"RemoteDelete",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("bool"),
	).Return(&YamsRepositoryError{})
	result := yamsService.RemoteDelete(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

func TestGetMaxConcurrency(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	yamsService := YamsService{YamsRepo: &mYamsRepo}
	expected := 1
	mYamsRepo.On("GetMaxConcurrency").Return(expected)
	result := yamsService.GetMaxConcurrency()
	assert.Equal(t, expected, result)
	mYamsRepo.AssertExpectations(t)
}

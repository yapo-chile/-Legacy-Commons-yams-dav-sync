package usecases

import (
	"testing"

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
	repo := SyncInteractor{YamsRepo: &mYamsRepo}

	input := domain.Image{
		Metadata: domain.ImageMetadata{
			ImageName: "name",
			Checksum:  "123",
		},
	}
	mYamsRepo.On("HeadImage", "name").
		Return(input.Metadata.Checksum, &YamsRepositoryError{})
	result := repo.ValidateChecksum(input)
	assert.Equal(t, true, result)
	mYamsRepo.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	repo := SyncInteractor{YamsRepo: &mYamsRepo}
	input := domain.Image{}
	mYamsRepo.On("PutImage", mock.AnythingOfType("domain.Image")).
		Return(&YamsRepositoryError{})
	result := repo.Send(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

func TestList(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	repo := SyncInteractor{YamsRepo: &mYamsRepo}
	mYamsRepo.On("GetImages").Return([]YamsObject{}, &YamsRepositoryError{})
	result, err := repo.List()
	assert.Equal(t, []YamsObject{}, result)
	assert.Equal(t, &YamsRepositoryError{}, err)
	mYamsRepo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	mYamsRepo := mockYamsRepo{}
	repo := SyncInteractor{YamsRepo: &mYamsRepo}
	input := "123.jpg"
	mYamsRepo.On(
		"DeleteImage",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("bool"),
	).Return(&YamsRepositoryError{})
	result := repo.Delete(input)
	assert.Equal(t, &YamsRepositoryError{}, result)
	mYamsRepo.AssertExpectations(t)
}

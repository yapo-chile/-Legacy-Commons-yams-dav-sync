package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
)

func TestNewLocalImageRepo(t *testing.T) {
	var fileSystemView FileSystemView
	imgRepo := &LocalImageRepo{
		fileSystemView: fileSystemView,
	}
	result := NewLocalImageRepo("", fileSystemView)
	assert.Equal(t, imgRepo, result)
}

func TestOpenFile(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}
	expected := &os.File{}
	mFileSystem.On("Open", mock.AnythingOfType("string")).Return(expected, nil)
	result, err := imgRepo.OpenFile("")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mFileSystem.AssertExpectations(t)
}

func TestGetImagePathTooShort(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}
	expected := domain.Image{}
	result, err := imgRepo.GetImage("")
	assert.Equal(t, expected, result)
	assert.Error(t, err)
	mFileSystem.AssertExpectations(t)
}

func TestGetImageOK(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	mFile := &mockFile{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	var int64Zero int64

	mFileSystem.On("Open", mock.AnythingOfType("string")).Return(mFile, nil)
	mFileSystem.On("Copy",
		mock.AnythingOfType("*md5.digest"),
		mock.AnythingOfType("*repository.mockFile")).Return(int64Zero, nil)
	mFileSystem.On("Name", mock.AnythingOfType("string")).Return("", nil)
	mFileSystem.On("Size", mock.AnythingOfType("string")).Return(int64Zero, nil)
	mFileSystem.On("ModTime", mock.AnythingOfType("string")).Return(time.Time{}, nil)

	mFile.On("Close").Return(nil).Once()

	expected := domain.Image{
		Metadata: domain.ImageMetadata{
			ModTime:  time.Time{},
			Checksum: "d41d8cd98f00b204e9800998ecf8427e",
		},
		FilePath: "fo/foto-sexy.jpg",
	}

	result, err := imgRepo.GetImage("foto-sexy.jpg")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mFileSystem.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestGetImageOpenError(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	mFile := &mockFile{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	mFileSystem.On("Open", mock.AnythingOfType("string")).Return(mFile, fmt.Errorf("err"))

	expected := domain.Image{}

	result, err := imgRepo.GetImage("foto-sexy.jpg")
	assert.Equal(t, expected, result)
	assert.Error(t, err)
	mFileSystem.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestGetImageCopyError(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	mFile := &mockFile{}
	imgRepo := &LocalImageRepo{
		path:           "",
		fileSystemView: mFileSystem,
	}

	var int64Zero int64

	mFileSystem.On("Open", mock.AnythingOfType("string")).Return(mFile, nil)
	mFileSystem.On("Copy",
		mock.AnythingOfType("*md5.digest"),
		mock.AnythingOfType("*repository.mockFile")).Return(int64Zero, fmt.Errorf("error"))

	mFile.On("Close").Return(nil).Once()

	expected := domain.Image{}

	result, err := imgRepo.GetImage("foto-sexy.jpg")
	assert.Equal(t, expected, result)
	assert.Error(t, err)
	mFileSystem.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

func TestGetImageListElement(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	expected := ""
	mFileSystem.On("Text").Return(expected, nil)

	result := imgRepo.GetImageListElement()
	assert.Equal(t, expected, result)
	mFileSystem.AssertExpectations(t)
}

func TestNextImageListElement(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	expected := true
	mFileSystem.On("Scan").Return(expected)

	result := imgRepo.NextImageListElement()
	assert.Equal(t, expected, result)
	mFileSystem.AssertExpectations(t)
}

func TestErrorScanningImageList(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	expected := fmt.Errorf("err")
	mFileSystem.On("Err").Return(expected)

	result := imgRepo.ErrorScanningImageList()
	assert.Equal(t, expected, result)
	mFileSystem.AssertExpectations(t)
}

func TestInitImageListScanner(t *testing.T) {
	mFileSystem := &mockFileSystemView{}
	mFile := &mockFile{}
	imgRepo := &LocalImageRepo{
		fileSystemView: mFileSystem,
	}

	mFileSystem.On("NewScanner",
		mock.AnythingOfType("*repository.mockFile")).Return()

	imgRepo.InitImageListScanner(mFile)
	mFileSystem.AssertExpectations(t)
	mFile.AssertExpectations(t)
}

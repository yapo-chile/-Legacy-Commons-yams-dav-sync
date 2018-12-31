package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewLastSyncRepo(t *testing.T) {
	var dbHandler DbHandler
	lastSyncRepo := &lastSyncRepo{
		db:          dbHandler,
		defaultDate: time.Time{},
	}
	result := NewLastSyncRepo(dbHandler, time.Time{})
	assert.Equal(t, lastSyncRepo, result)
}

func TestGetLastSync(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: time.Time{},
	}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	expected := time.Time{}
	result := lastSyncRepo.GetLastSync()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetLastSyncErrQuery(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: date,
	}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	expected := lastSyncRepo.defaultDate
	result := lastSyncRepo.GetLastSync()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetLastSyncErrScan(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	layout := "20060102T150405"
	date, _ := time.Parse(layout, "20180102T150405")
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: date,
	}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	expected := lastSyncRepo.defaultDate
	result := lastSyncRepo.GetLastSync()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetLastSync(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)

	err := lastSyncRepo.SetLastSync("123")

	assert.Nil(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetLastSyncQueryError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	err := lastSyncRepo.SetLastSync("123")

	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetLastSyncMarkError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	err := lastSyncRepo.SetLastSync("")

	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

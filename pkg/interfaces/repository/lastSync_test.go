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
	result := NewLastSyncRepo(dbHandler, "", time.Time{})
	assert.Equal(t, lastSyncRepo, result)
}

func TestGetLastSynchronizationMark(t *testing.T) {
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
	result := lastSyncRepo.GetLastSynchronizationMark()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetLastSynchronizationMarkErrQuery(t *testing.T) {
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
	result := lastSyncRepo.GetLastSynchronizationMark()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetLastSynchronizationMarkErrScan(t *testing.T) {
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
	result := lastSyncRepo.GetLastSynchronizationMark()

	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetLastSynchronizationMark(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	mDbHandler.On("Insert", mock.AnythingOfType("string")).Return(nil)

	err := lastSyncRepo.SetLastSynchronizationMark(time.Now())

	assert.NoError(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetLastSynchronizationMarkQueryError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	mDbHandler.On("Insert", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	err := lastSyncRepo.SetLastSynchronizationMark(time.Now())

	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

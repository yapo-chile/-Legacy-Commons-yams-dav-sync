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

	mDbHandler.On("Query", mock.AnythingOfType("string"), []interface{}(nil)).Return(mResult, nil)
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

	mDbHandler.On("Query", mock.AnythingOfType("string"), []interface{}(nil)).Return(mResult, fmt.Errorf("err"))
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

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
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
	lastSyncRepo := &lastSyncRepo{
		db: mDbHandler,
	}

	mDbHandler.On("Insert", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(nil)

	err := lastSyncRepo.SetLastSynchronizationMark(time.Now())

	assert.NoError(t, err)
	mDbHandler.AssertExpectations(t)
}

func TestGet(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: time.Time{},
	}

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	mResult.On("Next").Return(false).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	expected := []string{""}
	result, err := lastSyncRepo.Get()
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: time.Time{},
	}

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	expected := []string{}
	result, err := lastSyncRepo.Get()
	assert.Error(t, err)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetQueryError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: time.Time{},
	}

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, fmt.Errorf("err"))

	expected := []string{}
	result, err := lastSyncRepo.Get()
	assert.Error(t, err)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestReset(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	lastSyncRepo := &lastSyncRepo{
		db:          mDbHandler,
		defaultDate: time.Time{},
	}

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)

	err := lastSyncRepo.Reset()
	assert.NoError(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

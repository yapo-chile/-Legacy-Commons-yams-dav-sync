package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewErrorControlRepo(t *testing.T) {
	var dbHandler DbHandler
	errorControlRepo := &errorControlRepo{
		db: dbHandler,
	}
	result := NewErrorControlRepo(dbHandler, 0)
	assert.Equal(t, errorControlRepo, result)
}

func TestGetSyncErrors(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	expected := []string{""}

	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	mResult.On("Next").Return(false).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	result, err := errCtrlRepo.GetSyncErrors(1, 1)

	assert.Equal(t, expected, result)
	assert.Nil(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetSyncErrorsCloseError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	_, err := errCtrlRepo.GetSyncErrors(1, 1)

	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetPagesQty(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	expected := 0

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	result := errCtrlRepo.GetPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)

}

func TestGetPagesQtyErr(t *testing.T) {
	errCtrlRepo := &errorControlRepo{
		resultsPerPage: 0,
	}
	expected := 0
	result := errCtrlRepo.GetPagesQty(expected)
	assert.Equal(t, expected, result)
}

func TestGetPagesQtyResultQueryError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)
	expected := 0

	result := errCtrlRepo.GetPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetPagesQtyResultScanError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	expected := 0

	mResult.On("Scan", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	result := errCtrlRepo.GetPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestDelSyncError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, nil)
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.DelSyncError("fotito.jpg")
	assert.Nil(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetErrorCounter(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.SetErrorCounter("fotito.jpg", 0)
	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestAddSyncError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.AddSyncError("fotito.jpg")
	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

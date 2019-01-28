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

func TestGetPreviousErrors(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	expected := []string{""}

	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	mResult.On("Next").Return(false).Once()

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	result, err := errCtrlRepo.GetPreviousErrors(1, 1)

	assert.Equal(t, expected, result)
	assert.NoError(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetPreviousErrorsCloseError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	_, err := errCtrlRepo.GetPreviousErrors(1, 1)

	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetErrorsPagesQty(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	expected := 0

	mResult.On("Scan", mock.AnythingOfType("string")).Return(nil)

	result := errCtrlRepo.GetErrorsPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)

}

func TestGetErrorsPagesQtyErr(t *testing.T) {
	errCtrlRepo := &errorControlRepo{
		resultsPerPage: 0,
	}
	expected := 0
	result := errCtrlRepo.GetErrorsPagesQty(expected)
	assert.Equal(t, expected, result)
}

func TestGetErrorsPagesQtyResultQueryError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)
	expected := 0

	result := errCtrlRepo.GetErrorsPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestGetErrorsPagesQtyResultScanError(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db:             mDbHandler,
		resultsPerPage: 1,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)
	mResult.On("Next").Return(true).Once()
	expected := 0

	mResult.On("Scan", mock.AnythingOfType("string")).Return(fmt.Errorf("err"))

	result := errCtrlRepo.GetErrorsPagesQty(expected)
	assert.Equal(t, expected, result)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestCleanErrorMarks(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, nil)
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.CleanErrorMarks("fotito.jpg")
	assert.NoError(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestSetErrorCounter(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.SetErrorCounter("fotito.jpg", 0)
	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

func TestIncreaseErrorCounter(t *testing.T) {
	mDbHandler := &mockDbHandler{}
	mResult := &mockResult{}
	errCtrlRepo := &errorControlRepo{
		db: mDbHandler,
	}
	mDbHandler.On("Query", mock.AnythingOfType("string"),
		mock.AnythingOfType("[]interface {}")).Return(mResult, fmt.Errorf("err"))
	mResult.On("Close").Return(nil)

	err := errCtrlRepo.IncreaseErrorCounter("fotito.jpg")
	assert.Error(t, err)
	mDbHandler.AssertExpectations(t)
	mResult.AssertExpectations(t)
}

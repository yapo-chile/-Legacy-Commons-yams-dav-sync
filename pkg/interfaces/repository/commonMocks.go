package repository

import (
	"io"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockHTTPHandler struct { // nolint: deadcode
	mock.Mock
}

func (m *mockHTTPHandler) Send(request HTTPRequest) (HTTPResponse, error) {
	args := m.Called(request)
	return args.Get(0).(HTTPResponse), args.Error(1)
}

func (m *mockHTTPHandler) NewRequest() HTTPRequest {
	args := m.Called()
	return args.Get(0).(HTTPRequest)
}

type mockRequest struct { // nolint: deadcode
	mock.Mock
}

func (m *mockRequest) SetMethod(method string) HTTPRequest {
	args := m.Called(method)
	return args.Get(0).(HTTPRequest)
}

func (m *mockRequest) GetMethod() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockRequest) SetPath(path string) HTTPRequest {
	args := m.Called(path)
	return args.Get(0).(HTTPRequest)
}

func (m *mockRequest) GetPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockRequest) SetHeaders(headers map[string]string) HTTPRequest {
	args := m.Called(headers)
	return args.Get(0).(HTTPRequest)
}

func (m *mockRequest) GetHeaders() map[string][]string {
	args := m.Called()
	return args.Get(0).(map[string][]string)
}

func (m *mockRequest) SetBody(body interface{}) HTTPRequest {
	args := m.Called()
	return args.Get(0).(HTTPRequest)
}

func (m *mockRequest) GetBody() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *mockRequest) SetQueryParams(queryParams map[string]string) HTTPRequest {
	args := m.Called(queryParams)
	return args.Get(0).(HTTPRequest)
}

func (m *mockRequest) GetQueryParams() map[string][]string {
	args := m.Called()
	return args.Get(0).(map[string][]string)
}

func (m *mockRequest) GetTimeOut() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockRequest) SetTimeOut(t int) HTTPRequest {
	args := m.Called(t)
	return args.Get(0).(HTTPRequest)

}

func (m *mockRequest) SetImgBody(body io.Reader) HTTPRequest {
	args := m.Called(body)
	return args.Get(0).(HTTPRequest)
}

package repository

import "io"

// HTTPRequest implements HTTP request operations
type HTTPRequest interface {
	GetMethod() string
	SetMethod(string) HTTPRequest
	GetPath() string
	SetPath(string) HTTPRequest
	GetBody() interface{}
	SetBody(interface{}) HTTPRequest
	GetHeaders() map[string][]string
	SetHeaders(map[string]string) HTTPRequest
	GetQueryParams() map[string][]string
	SetQueryParams(map[string]string) HTTPRequest
	SetImgBody(body io.Reader) HTTPRequest
}

// HTTPHandler implements HTTP handler operations
type HTTPHandler interface {
	Send(HTTPRequest) (body interface{}, code int, err error)
	NewRequest() HTTPRequest
}

// HTTPRepository struct that contains httpHandler and Path to connect with
// external repositories
type HTTPRepository struct {
	Handler HTTPHandler
	Path    string
	Headers map[string]string
}

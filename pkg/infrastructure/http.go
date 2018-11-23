package infrastructure

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Yapo/logger"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// errorCodes codes that microservice considered as error code
var errorCodes = map[int]string{
	http.StatusBadRequest:          "Bad request",
	http.StatusInternalServerError: "Internal server error",
}

// HTTPHandler struct to implements http repository operations
type HTTPHandler struct{}

// NewHTTPHandler will create a new instance of a custom http request handler
func NewHTTPHandler() repository.HTTPHandler {
	return &HTTPHandler{}
}

// Send will execute the sending of a http request
// a custom http client has been made to add a request timeout of 10 seconds
func (*HTTPHandler) Send(req repository.HTTPRequest) (body interface{}, code int, err error) {
	logger.Debug("HTTP - %s - Sending HTTP request to: %+v", req.GetMethod(), req.GetPath())

	// this makes a custom http client with a timeout in secs for each request
	var httpClient = &http.Client{
		Timeout: time.Second * time.Duration(10),
	}
	resp, err := httpClient.Do(&req.(*request).innerRequest)
	if err != nil {
		logger.Error("HTTP - %s - Error sending HTTP request: %+v", req.GetMethod(), err)
		return nil, resp.StatusCode, fmt.Errorf("Found error: %+v", err)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if val, ok := errorCodes[resp.StatusCode]; ok {
		logger.Error("HTTP - %s - Received an error response: %+v", req.GetMethod(), val)
		return nil, resp.StatusCode, fmt.Errorf("%s", response)
	}
	if err != nil {
		logger.Error("HTTP - %s - Error reading response: %+v", req.GetMethod(), err)
	}
	return string(response), resp.StatusCode, nil
}

// request is a custom golang http.Request
type request struct {
	innerRequest http.Request
	body         interface{}
}

// NewRequest returns an initialized struct that can be used to make a http request
func (*HTTPHandler) NewRequest() repository.HTTPRequest {
	return &request{
		innerRequest: http.Request{
			Header: make(http.Header),
		},
	}
}

// SetMethod sets the HTTP method to be used, like GET, POST, PUT, etc
func (r *request) SetMethod(method string) repository.HTTPRequest {
	r.innerRequest.Method = method
	return r
}

// GetMethod retrieves the actual HTTP method
func (r *request) GetMethod() string {
	return r.innerRequest.Method
}

// SetPath sets the url path that will be requested
func (r *request) SetPath(path string) repository.HTTPRequest {
	url, err := url.Parse(path)
	r.innerRequest.URL = url
	if err != nil {
		logger.Error("HTTP - there was an error setting the request path: %+v", err)
	}
	return r
}

// GetPath retrieves the actual url path
func (r request) GetPath() string {
	return r.innerRequest.URL.String()
}

// SetHeaders will set custom headers to the request
func (r *request) SetHeaders(headers map[string]string) repository.HTTPRequest {
	for header, value := range headers {
		r.innerRequest.Header.Set(header, value)
	}
	return r
}

// GetHeaders will retrieve the custom headers of the request
func (r *request) GetHeaders() map[string][]string {
	return r.innerRequest.Header
}

// SetBody will set a custom body to the request, this body is the json representation of an interface{}
// this method will also set the custom header Content-type to application-json
// and will save the original body
func (r *request) SetBody(body interface{}) repository.HTTPRequest {
	var reader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			logger.Error("HTTP - Error parsing request data to json: %+v", err)
		}
		reader = strings.NewReader(string(jsonBody))
	}
	// if SetBody is called then we add the Content-type header as a default
	r.SetHeaders(map[string]string{"Content-type": "application/json"})
	r.innerRequest.Body = ioutil.NopCloser(reader)

	// this will be usefull if we need to call GetBody(...)
	r.body = body
	return r
}

// SetImgBody will set a custom img body to the request.
// this method will also set the custom header Content-type to images/jpg
func (r *request) SetImgBody(body io.Reader) repository.HTTPRequest {
	r.SetHeaders(map[string]string{"Content-type": "images/jpg"})
	r.innerRequest.Body = ioutil.NopCloser(body)
	r.body = body
	return r
}

// GetBody retrieves the original interface{} set on this request
// so after calling this methos you should be able to assert it to its original type
func (r *request) GetBody() interface{} {
	return r.body
}

// SetQueryParams will set custom query parameters to the request
func (r *request) SetQueryParams(queryParams map[string]string) repository.HTTPRequest {
	q := r.innerRequest.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	r.innerRequest.URL.RawQuery = q.Encode()
	return r
}

// GetQueryParams will retrieve the query parameters of the request
func (r *request) GetQueryParams() map[string][]string {
	return r.innerRequest.URL.Query()
}

package repository

import (
	"io"
	"net/http"
	"time"

	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/interfaces"
	"github.mpi-internal.com/Yapo/yams-dav-sync/pkg/usecases"
)

// DbHandler represents a database connection handler
// it provides basic database capabilities
// after its use, the connection with the database must be closed
type DbHandler interface {
	io.Closer
	Insert(statement string, params ...interface{}) error
	Update(statement string, params ...interface{}) error
	Query(statement string, params ...interface{}) (DbResult, error)
}

// DbRepo contains an instance of a DBHandler
type DbRepo struct {
	Handler DbHandler
}

// DbResult represents a database query result rows
// after its use, the Close() method must be invoked
// to ensure that the database connection used to perform the query
// returns to the connection pool to be used again
type DbResult interface {
	Scan(dest ...interface{}) error
	Next() bool
	io.Closer
}

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
	GetTimeOut() time.Duration
	SetTimeOut(timeout int) HTTPRequest
}

// HTTPHandler implements HTTP handler operations
type HTTPHandler interface {
	Send(HTTPRequest) (HTTPResponse, error)
	NewRequest() HTTPRequest
}

// HTTPRepository struct that contains httpHandler and Path to connect with
// external repositories
type HTTPRepository struct {
	Handler HTTPHandler
	Path    string
	Headers map[string]string
	TimeOut int
}

// HTTPResponse struct that contains http response of
type HTTPResponse struct {
	Body    interface{}
	Code    int
	Headers http.Header
}

// FileSystemView allows FileSystem's operations to view elements in local storage
type FileSystemView interface {
	Open(name string) (usecases.File, error)
	NewScanner(usecases.File) interfaces.Scanner
	Copy(dst io.Writer, src io.Reader) error
	Info(name string) (FileInfo, error)
}

// FileInfo allows operations to get file information in local storage
type FileInfo interface {
	ModTime() time.Time
	Name() string
	Size() int64
}

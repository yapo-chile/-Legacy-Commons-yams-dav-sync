package repository

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type mockSigner struct {
	mock.Mock
}

func (m *mockSigner) GenerateTokenString(claims jwt.Claims) string {
	args := m.Called(claims)
	return args.String(0)
}

type MockYamsRepoLogger struct {
	mock.Mock
}

func (m *MockYamsRepoLogger) LogRequestURI(url string) {
	m.Called(url)
}

func (m *MockYamsRepoLogger) LogStatus(statusCode int) {
	m.Called(statusCode)
}

func (m *MockYamsRepoLogger) LogResponse(body string, err error) {
	m.Called(body, err)
}

func (m *MockYamsRepoLogger) LogCannotDecodeErrorMessage(err error) {
	m.Called(err)
}

func TestNewYamsRepository(t *testing.T) {
	var jwtSigner Signer
	var logger YamsRepositoryLogger
	var http HTTPHandler
	yamsRepo := YamsRepository{
		jwtSigner:   jwtSigner,
		mgmtURL:     "url",
		accessKeyID: "key",
		tenantID:    "tentantID",
		domainID:    "domainID",
		bucketID:    "bucketID",
		logger:      logger,
		http: &HTTPRepository{
			Handler: http,
			TimeOut: 0,
		},
		maxConcurrentThreads: 100,
	}
	result := NewYamsRepository(yamsRepo.jwtSigner, yamsRepo.mgmtURL, yamsRepo.accessKeyID,
		yamsRepo.tenantID, yamsRepo.domainID, yamsRepo.bucketID, nil, yamsRepo.logger, http, 0,
		"", "", yamsRepo.maxConcurrentThreads)
	assert.Equal(t, &yamsRepo, result)
}

func TestGetMaxConcurrency(t *testing.T) {
	yamsRepo := YamsRepository{
		maxConcurrentThreads: 100,
	}
	result := yamsRepo.GetMaxConcurrency()
	assert.Equal(t, yamsRepo.maxConcurrentThreads, result)
}

func TestGetDomain(t *testing.T) {
	yamsRepo := YamsRepository{}
	result := yamsRepo.GetMaxConcurrency()
	assert.Equal(t, yamsRepo.maxConcurrentThreads, result)
}

func TestGetDomains(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
	}

	mHandler.On("NewRequest").Return(&mRequest, nil).Once()
	response := HTTPResponse{
		Code: 200,
		Body: `domains`,
	}

	mRequest.On("SetMethod", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetPath", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetQueryParams", mock.AnythingOfType("map[string]string")).Return(&mRequest)

	mLogger.On("LogRequestURI", mock.AnythingOfType("string"))
	mLogger.On("LogStatus", mock.AnythingOfType("int"))
	mLogger.On(
		"LogResponse",
		mock.AnythingOfType("string"),
		nil,
	)

	mHandler.On("Send", &mRequest).Return(response, nil).Once()

	mSigner.On("GenerateTokenString", mock.AnythingOfType("MyCustomClaims")).Return("claims")

	domains := yamsRepo.GetDomains()

	assert.Equal(t, response.Body, domains)

	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mRequest.AssertExpectations(t)
	mHandler.AssertExpectations(t)
}

type mockFileSystemView struct {
	mock.Mock
}

func (m *mockFileSystemView) Open(name string) (usecases.File, error) {
	args := m.Called(name)
	return args.Get(0).(usecases.File), args.Error(1)
}

func (m *mockFileSystemView) NewScanner(file usecases.File) interfaces.Scanner {
	args := m.Called(file)
	return args.Get(0).(interfaces.Scanner)
}

func (m *mockFileSystemView) Copy(dst io.Writer, src io.Reader) error {
	args := m.Called(dst, src)
	return args.Error(0)
}

func (m *mockFileSystemView) Info(filePath string) (FileInfo, error) {
	args := m.Called(filePath)
	return args.Get(0).(FileInfo), args.Error(1)
}

type mockFileInfo struct {
	mock.Mock
}

func (m *mockFileInfo) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFileInfo) Size() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *mockFileInfo) ModTime() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

type mockFile struct {
	mock.Mock
}

func (m *mockFile) Close() (err error) {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFile) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func TestSendErrorOpeningFile(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}
	mFileSystemView := mockFileSystemView{}
	mFile := mockFile{}

	localImageRepo := NewLocalImageRepo("", &mFileSystemView)
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
		localImageRepo: localImageRepo,
	}

	mSigner.On("GenerateTokenString", mock.AnythingOfType("PutClaims")).
		Return("claims")

	mFileSystemView.On("Open", mock.AnythingOfType("string")).
		Return(&mFile, fmt.Errorf("err"))

	remoteChecksum, resp := yamsRepo.Send(domain.Image{})

	assert.Equal(t, usecases.ErrYamsImage, resp)
	assert.Equal(t, "", remoteChecksum)

	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mHandler.AssertExpectations(t)
	mRequest.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mFileSystemView.AssertExpectations(t)
}

func TestSend(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}
	mFileSystemView := mockFileSystemView{}
	mFile := mockFile{}

	localImageRepo := NewLocalImageRepo("", &mFileSystemView)
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
		localImageRepo: localImageRepo,
	}

	mHandler.On("NewRequest").Return(&mRequest, nil)

	mFileSystemView.On("Open", mock.AnythingOfType("string")).Return(&mFile, nil)

	mRequest.On("SetMethod", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetPath", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetQueryParams", mock.AnythingOfType("map[string]string")).Return(&mRequest)
	mRequest.On("SetImgBody", mock.AnythingOfType("*repository.mockFile")).Return(&mRequest)
	mRequest.On("SetTimeOut", mock.AnythingOfType("int")).Return(&mRequest)
	mRequest.On("SetHeaders", mock.AnythingOfType("map[string]string")).Return(&mRequest)

	mFile.On("Close").Return(nil)
	mLogger.On("LogStatus", mock.AnythingOfType("int"))
	mLogger.On("LogResponse", mock.AnythingOfType("string"), nil)
	mSigner.On("GenerateTokenString", mock.AnythingOfType("PutClaims")).Return("claims")
	mLogger.On("LogCannotDecodeErrorMessage", mock.AnythingOfType("*json.SyntaxError"))
	expected := ""
	for cases := 0; cases < 7; cases++ {
		switch cases {
		case 0: // everything OK
			response := HTTPResponse{
				Code: 200,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Nil(t, resp)

		case 1: // 400 Internal error
			response := HTTPResponse{
				Code: 400,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsInternal, resp)

		case 2: // 403 Unauthorized error
			response := HTTPResponse{
				Code: 403,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsUnauthorized, resp)
		case 3: // 404 Bucket Not Found error
			response := HTTPResponse{
				Code: 404,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsBucketNotFound, resp)
		case 4: // 409 object duplicated error
			response := HTTPResponse{
				Code: 409,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsDuplicate, resp)
		case 5: // 500 Internal Server error
			response := HTTPResponse{
				Code: 500,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsInternal, resp)
		case 6: // 503 Yams internal error
			response := HTTPResponse{
				Code: 503,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			remoteChecksum, resp := yamsRepo.Send(domain.Image{})
			assert.Equal(t, expected, remoteChecksum)
			assert.Equal(t, usecases.ErrYamsInternal, resp)
		}
	}

	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mHandler.AssertExpectations(t)
	mRequest.AssertExpectations(t)
	mFile.AssertExpectations(t)
	mFileSystemView.AssertExpectations(t)
}

func TestRemoteDelete(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}

	localImageRepo := NewLocalImageRepo("", nil)
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
		localImageRepo: localImageRepo,
	}

	mHandler.On("NewRequest").Return(&mRequest, nil)

	mRequest.On("SetMethod", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetPath", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetQueryParams", mock.AnythingOfType("map[string]string")).Return(&mRequest)
	mRequest.On("SetTimeOut", mock.AnythingOfType("int")).Return(&mRequest)

	mLogger.On("LogStatus", mock.AnythingOfType("int"))
	mLogger.On("LogRequestURI", mock.AnythingOfType("string"))
	mLogger.On("LogResponse", mock.AnythingOfType("string"), nil)
	mSigner.On("GenerateTokenString", mock.AnythingOfType("DeleteClaims")).Return("claims")

	for cases := 0; cases < 7; cases++ {
		switch cases {
		case 0: // everything OK
			response := HTTPResponse{
				Code: 202,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Nil(t, resp)
		case 1: // 400 yams internal error
			response := HTTPResponse{
				Code: 400,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsInternal)
		case 2: // 403 yams Unauthorized error
			response := HTTPResponse{
				Code: 403,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsUnauthorized)
		case 3: // 404 object not found error
			response := HTTPResponse{
				Code: 404,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsObjectNotFound)
		case 4: // 500 server error
			response := HTTPResponse{
				Code: 500,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsInternal)
		case 5: // 503 Service temporarily unavailable
			response := HTTPResponse{
				Code: 503,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsInternal)
		default: // Unknown error
			response := HTTPResponse{
				Code: 999,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp := yamsRepo.RemoteDelete("foto-sexy.jpg", domain.YAMSForceRemoval)
			assert.Equal(t, resp, usecases.ErrYamsInternal)
		}
	}
	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mHandler.AssertExpectations(t)
	mRequest.AssertExpectations(t)
}

func TestGetRemoteChecksum(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}

	localImageRepo := NewLocalImageRepo("", nil)
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
		localImageRepo: localImageRepo,
	}

	mHandler.On("NewRequest").Return(&mRequest, nil)

	mRequest.On("SetMethod", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetPath", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetQueryParams", mock.AnythingOfType("map[string]string")).Return(&mRequest)
	mRequest.On("SetTimeOut", mock.AnythingOfType("int")).Return(&mRequest)

	mLogger.On("LogStatus", mock.AnythingOfType("int"))
	mLogger.On("LogRequestURI", mock.AnythingOfType("string"))
	mLogger.On("LogResponse", mock.AnythingOfType("string"), nil)
	mSigner.On("GenerateTokenString", mock.AnythingOfType("InfoClaims")).Return("claims")

	for cases := 0; cases < 5; cases++ {
		switch cases {
		case 0: // everything OK
			expected := "algo en md5"
			response := HTTPResponse{
				Code: 200,
				Headers: http.Header{
					"Content-Md5": []string{expected},
				},
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp, err := yamsRepo.GetRemoteChecksum("foto-sexy.jpg")
			assert.Nil(t, err)
			assert.Equal(t, expected, resp)
		case 1: // 400 yams internal error
			response := HTTPResponse{
				Code: 404,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.GetRemoteChecksum("foto-sexy.jpg")
			assert.Equal(t, usecases.ErrYamsObjectNotFound, err)

		case 2: // 500 server error
			response := HTTPResponse{
				Code: 500,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.GetRemoteChecksum("foto-sexy.jpg")
			assert.Equal(t, usecases.ErrYamsInternal, err)
		case 3: // 503 Service temporarily unavailable
			response := HTTPResponse{
				Code: 503,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.GetRemoteChecksum("foto-sexy.jpg")
			assert.Equal(t, usecases.ErrYamsInternal, err)
		default: // Unknown error
			response := HTTPResponse{
				Code: 999,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.GetRemoteChecksum("foto-sexy.jpg")
			assert.Equal(t, usecases.ErrYamsInternal, err)
		}
	}
	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mHandler.AssertExpectations(t)
	mRequest.AssertExpectations(t)
}

func TestGetLocalImages(t *testing.T) {
	mLogger := MockYamsRepoLogger{}
	mSigner := mockSigner{}
	mHandler := mockHTTPHandler{}
	mRequest := mockRequest{}

	localImageRepo := NewLocalImageRepo("", nil)
	yamsRepo := YamsRepository{
		jwtSigner: &mSigner,
		logger:    &mLogger,
		http: &HTTPRepository{
			Handler: &mHandler,
		},
		localImageRepo: localImageRepo,
	}

	mHandler.On("NewRequest").Return(&mRequest, nil)

	mRequest.On("SetMethod", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetPath", mock.AnythingOfType("string")).Return(&mRequest)
	mRequest.On("SetQueryParams", mock.AnythingOfType("map[string]string")).Return(&mRequest)
	mRequest.On("SetTimeOut", mock.AnythingOfType("int")).Return(&mRequest)

	mLogger.On("LogRequestURI", mock.AnythingOfType("string"))
	mLogger.On("LogResponse", mock.AnythingOfType("string"), nil)
	mSigner.On("GenerateTokenString", mock.AnythingOfType("InfoClaims")).Return("claims")

	body := []byte(`{"objects":[{"object_id":"123","md5":"algo en md5",` +
		`"size":1, "last_modified":1}], "continuation_token":""}`)

	for cases := 0; cases < 6; cases++ {
		switch cases {
		case 0: // everything OK
			expected := []usecases.YamsObject{
				usecases.YamsObject{
					ID:           "123",
					Md5:          "algo en md5",
					Size:         1,
					LastModified: 1,
				},
			}
			response := HTTPResponse{
				Code: 200,
				Body: body,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			resp, err := yamsRepo.List()
			assert.Nil(t, err)
			assert.Equal(t, expected, resp)
		case 1: // 404 yams objects not found
			response := HTTPResponse{
				Code: 404,
				Body: body,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.List()
			assert.Equal(t, usecases.ErrYamsObjectNotFound, err)

		case 2: // 500 object not found error
			response := HTTPResponse{
				Code: 500,
				Body: body,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.List()
			assert.Equal(t, usecases.ErrYamsInternal, err)
		case 3: // 503 Service temporarily unavailable
			response := HTTPResponse{
				Code: 503,
				Body: body,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.List()
			assert.Equal(t, usecases.ErrYamsInternal, err)
		case 4: // Unmarshal error
			response := HTTPResponse{
				Body: "+++++++",
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.List()
			assert.Equal(t, usecases.ErrYamsInternal, err)
		default: // Unknown error
			response := HTTPResponse{
				Code: 999,
				Body: body,
			}
			mHandler.On("Send", &mRequest).Return(response, nil).Once()
			_, err := yamsRepo.List()
			assert.Equal(t, usecases.ErrYamsInternal, err)
		}
	}
	mLogger.AssertExpectations(t)
	mSigner.AssertExpectations(t)
	mHandler.AssertExpectations(t)
	mRequest.AssertExpectations(t)
}

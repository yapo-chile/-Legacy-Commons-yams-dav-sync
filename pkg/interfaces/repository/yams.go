package repository

import (
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

// YamsRepository is yams bucket representation that allows operations
// execution using http requests
type YamsRepository struct {
	// maxConcurrentThreads max number of concurrent connection to yams
	maxConcurrentThreads int
	// jwtSigner validates each request with jwt signature
	jwtSigner Signer
	// mgtURL contains the url of yams management server
	mgmtURL string
	// accessKeyID is the user accesskey connected to yams server
	accessKeyID string
	// tenantID is the yams tentant environment that contains domains
	tenantID string
	// domainID is the yams domain that contains buckets
	domainID string
	// bucketID is the yams bucket that contains images
	bucketID string
	// http is the http client to connect to yams using http protocol
	http *HTTPRepository
	// localImageRepo  repo to execute operations in local storage
	localImageRepo interfaces.LocalImage
	// logger logs yams repository events
	logger YamsRepositoryLogger
}

// Signer allows methods to validate each request to yams server
type Signer interface {
	GenerateTokenString(claims jwt.Claims) string
}

// NewYamsRepository creates a new instance of YamsRepository
func NewYamsRepository(jwtSigner Signer, mgmtURL, accessKeyID, tenantID,
	domainID, bucketID string, localImageRepo interfaces.LocalImage, logger YamsRepositoryLogger, handler HTTPHandler,
	timeOut int, maxConcurrentThreads int) *YamsRepository {
	return &YamsRepository{
		jwtSigner:   jwtSigner,
		mgmtURL:     mgmtURL,
		accessKeyID: accessKeyID,
		tenantID:    tenantID,
		domainID:    domainID,
		bucketID:    bucketID,
		logger:      logger,
		http: &HTTPRepository{
			Handler: handler,
			TimeOut: timeOut,
		},
		maxConcurrentThreads: maxConcurrentThreads,
		localImageRepo:       localImageRepo,
	}
}

// YamsRepositoryLogger allows methods to log yams repository events
type YamsRepositoryLogger interface {
	LogRequestURI(url string)
	LogStatus(statusCode int)
	LogResponse(body string, err error)
}

// GetMaxConcurrency gets the max number of concurrent connections to yams
func (repo *YamsRepository) GetMaxConcurrency() int {
	return repo.maxConcurrentThreads
}

// GetDomains gets domains from yams, domains belongs to the repo tenant
func (repo *YamsRepository) GetDomains() string {

	type MyCustomClaims struct {
		jwt.StandardClaims
		Rqs string `json:"rqs"`
	}

	path := "/tenants/" + repo.tenantID + "/domains"

	// Create the Claims
	claims := MyCustomClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"GET\\" + path,
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := repo.mgmtURL + path
	repo.logger.LogRequestURI(requestURI)

	queryParams := map[string]string{
		"jwt":         tokenString,
		"AccessKeyId": repo.accessKeyID,
	}

	request := repo.http.Handler.
		NewRequest().
		SetMethod("GET").
		SetPath(requestURI).
		SetQueryParams(queryParams)
	resp, err := repo.http.Handler.Send(request)
	repo.logger.LogStatus(resp.Code)
	domains := fmt.Sprintf("%s", resp.Body)

	repo.logger.LogResponse(domains, err)

	return domains
}

// Send puts a image in yams repository
func (repo *YamsRepository) Send(image domain.Image) *usecases.YamsRepositoryError {
	type PutMetadata struct {
		ObjectID string `json:"oid"`
	}

	type PutClaims struct {
		jwt.StandardClaims
		Rqs      string      `json:"rqs"`
		Metadata PutMetadata `json:"metadata"`
	}
	path := "/tenants/" + repo.tenantID +
		"/domains/" + repo.domainID +
		"/buckets/" + repo.bucketID +
		"/objects"
	// Create the Claims
	claims := PutClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"POST\\" + path,
		PutMetadata{
			ObjectID: image.Metadata.ImageName,
		},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := repo.mgmtURL + path

	queryParams := map[string]string{
		"jwt":         tokenString,
		"AccessKeyId": repo.accessKeyID,
	}

	imageFile, err := repo.localImageRepo.OpenFile(image.FilePath)
	if err != nil {
		return usecases.ErrYamsImage
	}
	defer imageFile.Close() // nolint

	request := repo.http.Handler.
		NewRequest().
		SetMethod("POST").
		SetPath(requestURI).
		SetImgBody(imageFile).
		SetQueryParams(queryParams).
		SetTimeOut(repo.http.TimeOut)

	resp, err := repo.http.Handler.Send(request)
	repo.logger.LogStatus(resp.Code)
	body := fmt.Sprintf("%s", resp.Body)
	repo.logger.LogResponse(body, err)

	switch resp.Code {
	case 400: // Bad Request
		return usecases.ErrYamsInternal
	case 403:
		return usecases.ErrYamsUnauthorized
	case 404:
		return usecases.ErrYamsBucketNotFound
	case 409:
		return usecases.ErrYamsDuplicate
	case 500: // Server error
		return usecases.ErrYamsInternal
	case 503: // Service temporarily unavailable
		return usecases.ErrYamsInternal
	}

	return nil
}

// RemoteDelete deletes a specific image of yams repository
func (repo *YamsRepository) RemoteDelete(imageName string, immediateRemoval bool) *usecases.YamsRepositoryError {

	type DeleteMetadata struct {
		ObjectID              string `json:"oid"`
		ForceImmediateRemoval bool   `json:"force"`
	}

	type DeleteClaims struct {
		jwt.StandardClaims
		Rqs      string         `json:"rqs"`
		Metadata DeleteMetadata `json:"metadata"`
	}

	path := "/tenants/" + repo.tenantID +
		"/domains/" + repo.domainID +
		"/buckets/" + repo.bucketID +
		"/objects/" + imageName

	// Create the Claims
	claims := DeleteClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"DELETE\\" + path,
		DeleteMetadata{
			ForceImmediateRemoval: immediateRemoval,
			ObjectID:              imageName,
		},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := repo.mgmtURL + path

	repo.logger.LogRequestURI(requestURI)

	queryParams := map[string]string{
		"jwt":         tokenString,
		"AccessKeyId": repo.accessKeyID,
	}

	request := repo.http.Handler.
		NewRequest().
		SetMethod("DELETE").
		SetPath(requestURI).
		SetQueryParams(queryParams).
		SetTimeOut(repo.http.TimeOut)

	resp, err := repo.http.Handler.Send(request)
	repo.logger.LogStatus(resp.Code)
	body := fmt.Sprintf("%s", resp.Body)

	repo.logger.LogResponse(body, err)

	switch resp.Code {
	case 202: // All good, object deleted
		return nil
	case 400: // Bad Request
		return usecases.ErrYamsInternal
	case 403:
		return usecases.ErrYamsUnauthorized
	case 404:
		return usecases.ErrYamsObjectNotFound
	case 500: // Server error
		return usecases.ErrYamsInternal
	case 503: // Service temporarily unavailable
		return usecases.ErrYamsInternal
	default: // Unknown error
		return usecases.ErrYamsInternal
	}
}

// GetRemoteChecksum gets an object metadata.
func (repo *YamsRepository) GetRemoteChecksum(imageName string) (string, *usecases.YamsRepositoryError) {
	type InfoClaims struct {
		jwt.StandardClaims
		Rqs string `json:"rqs"`
	}

	path := "/tenants/" + repo.tenantID +
		"/domains/" + repo.domainID +
		"/buckets/" + repo.bucketID +
		"/objects/" + imageName

	// Create the Claims
	claims := InfoClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"HEAD\\" + path,
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := repo.mgmtURL + path

	repo.logger.LogRequestURI(requestURI)

	queryParams := map[string]string{
		"jwt":         tokenString,
		"AccessKeyId": repo.accessKeyID,
	}

	request := repo.http.Handler.
		NewRequest().
		SetMethod("HEAD").
		SetPath(requestURI).
		SetQueryParams(queryParams).
		SetTimeOut(repo.http.TimeOut)

	resp, err := repo.http.Handler.Send(request)
	repo.logger.LogStatus(resp.Code)
	body := fmt.Sprintf("%s", resp.Body)

	hashResponse := resp.Headers.Get("Content-Md5")

	repo.logger.LogResponse(body, err)

	switch resp.Code {
	case 200: // Headers are set and returned
		return hashResponse, nil
	case 404:
		return hashResponse, usecases.ErrYamsObjectNotFound
	case 500: // Server error
		return hashResponse, usecases.ErrYamsInternal
	case 503: // Service temporarily unavailable
		return hashResponse, usecases.ErrYamsInternal
	default: // Unkown error
		return hashResponse, usecases.ErrYamsInternal
	}
}

// List gets a list of available images in yams repository
func (repo *YamsRepository) List() ([]usecases.YamsObject, *usecases.YamsRepositoryError) {

	type InfoClaims struct {
		jwt.StandardClaims
		Rqs string `json:"rqs"`
	}

	path := "/tenants/" + repo.tenantID +
		"/domains/" + repo.domainID +
		"/buckets/" + repo.bucketID +
		"/objects"

	// Create the Claims
	claims := InfoClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"GET\\" + path,
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := repo.mgmtURL + path

	repo.logger.LogRequestURI(requestURI)

	queryParams := map[string]string{
		"jwt":         tokenString,
		"AccessKeyId": repo.accessKeyID,
	}

	request := repo.http.Handler.
		NewRequest().
		SetMethod("GET").
		SetPath(requestURI).
		SetQueryParams(queryParams).
		SetTimeOut(repo.http.TimeOut)
	resp, err := repo.http.Handler.Send(request)

	body := fmt.Sprintf("%s", resp.Body)

	repo.logger.LogResponse(body, err)

	var response usecases.YamsGetResponse
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return nil, usecases.ErrYamsInternal
	}
	switch resp.Code {
	case 200: // Headers are set and returned
		return response.Images, nil
	case 404:
		return nil, usecases.ErrYamsObjectNotFound
	case 500: // Server error
		return nil, usecases.ErrYamsInternal
	case 503: // Service temporarily unavailable
		return nil, usecases.ErrYamsInternal
	default: // Unkown error
		return nil, usecases.ErrYamsInternal
	}
}

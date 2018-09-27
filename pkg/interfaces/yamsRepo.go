package interfaces

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/usecases"
)

type YamsRepository struct {
	Debug       bool
	jwtSigner   infra.JWTSigner
	mgmtURL     string
	accessKeyID string
	tenantID    string
	domainID    string
	bucketID    string
	httpClient  *http.Client
}

func NewYamsRepository(jwtSigner infra.JWTSigner, mgmtURL, accessKeyID, tenantID, domainID, bucketID string) *YamsRepository {
	yamsRepo := &YamsRepository{
		jwtSigner:   jwtSigner,
		mgmtURL:     mgmtURL,
		accessKeyID: accessKeyID,
		tenantID:    tenantID,
		domainID:    domainID,
		bucketID:    bucketID,
	}

	// Disable debug by default
	yamsRepo.Debug = false

	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	yamsRepo.httpClient = &http.Client{Transport: tr}

	return yamsRepo
}

func (repo *YamsRepository) GetDomains() string {

	type Metadata struct{}

	type MyCustomClaims struct {
		jwt.StandardClaims
		Rqs      string   `json:"rqs"`
		Metadata Metadata `json:"metadata"`
	}

	// Create the Claims
	claims := MyCustomClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		"GET\\/tenants/" + repo.tenantID + "/domains",
		Metadata{},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := "https://" + repo.mgmtURL + "/api/v1/tenants/" + repo.tenantID + "/domains?jwt=" + tokenString + "&AccessKeyId=" + repo.accessKeyID
	if repo.Debug {
		fmt.Println(requestURI)
	}

	resp, err := http.Get(requestURI)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if repo.Debug {
		fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	}

	return ""
}

func (repo *YamsRepository) PutImage(image domain.Image) *usecases.YamsRepositoryError {

	type PutMetadata struct {
		ObjectID string `json:"oid"`
	}

	type PutClaims struct {
		jwt.StandardClaims
		Rqs      string      `json:"rqs"`
		Metadata PutMetadata `json:"metadata"`
	}

	// Create the Claims
	claims := PutClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		stringConcat("POST\\/tenants/", repo.tenantID, "/domains/", repo.domainID, "/buckets/", repo.bucketID, "/objects"),
		PutMetadata{
			ObjectID: image.Metadata.ImageName,
		},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := stringConcat("https://", repo.mgmtURL, "/api/v1/tenants/", repo.tenantID, "/domains/", repo.domainID,
		"/buckets/", repo.bucketID, "/objects?jwt=", tokenString, "&AccessKeyId=", repo.accessKeyID)

	if repo.Debug {
		fmt.Println(requestURI)
	}

	imageFile, err := os.Open(image.FilePath)
	if err != nil {
		return usecases.ErrYamsImage
	}
	defer imageFile.Close()

	resp, err := repo.httpClient.Post(requestURI, "images/jpg", imageFile)
	if err != nil {
		return usecases.ErrYamsConnection
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if repo.Debug {
		fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	}

	switch resp.StatusCode {
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

func (repo *YamsRepository) DeleteImage(imageName string, immediateRemoval bool) *usecases.YamsRepositoryError {

	type DeleteMetadata struct {
		ObjectID              string `json:"oid"`
		ForceImmediateRemoval bool   `json:"force"`
	}

	type DeleteClaims struct {
		jwt.StandardClaims
		Rqs      string         `json:"rqs"`
		Metadata DeleteMetadata `json:"metadata"`
	}

	urlPath := stringConcat("/tenants/", repo.tenantID, "/domains/", repo.domainID, "/buckets/", repo.bucketID, "/objects/", imageName)

	// Create the Claims
	claims := DeleteClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		stringConcat("DELETE\\", urlPath),
		DeleteMetadata{
			ForceImmediateRemoval: immediateRemoval,
			ObjectID:              imageName,
		},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := stringConcat("https://", repo.mgmtURL, "/api/v1", urlPath, "?jwt=", tokenString, "&AccessKeyId=", repo.accessKeyID)

	if repo.Debug {
		fmt.Println(requestURI)
	}

	req, err := http.NewRequest("DELETE", requestURI, nil)
	if err != nil {
		return usecases.ErrYamsConnection
	}

	resp, err := repo.httpClient.Do(req)
	if err != nil {
		return usecases.ErrYamsConnection
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if repo.Debug {
		fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	}

	switch resp.StatusCode {
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
	default: // Unkown error
		return usecases.ErrYamsInternal
	}
}

func (repo *YamsRepository) HeadImage(imageName string) *usecases.YamsRepositoryError {

	type InfoClaims struct {
		jwt.StandardClaims
		Rqs string `json:"rqs"`
	}

	urlPath := stringConcat("/tenants/", repo.tenantID, "/domains/", repo.domainID, "/buckets/", repo.bucketID, "/objects/", imageName)

	// Create the Claims
	claims := InfoClaims{
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		stringConcat("HEAD\\", urlPath),
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := stringConcat("https://", repo.mgmtURL, "/api/v1", urlPath, "?jwt=", tokenString, "&AccessKeyId=", repo.accessKeyID)

	if repo.Debug {
		fmt.Println(requestURI)
	}

	req, err := http.NewRequest("HEAD", requestURI, nil)
	if err != nil {
		return usecases.ErrYamsConnection
	}

	resp, err := repo.httpClient.Do(req)
	if err != nil {
		return usecases.ErrYamsConnection
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if repo.Debug {
		fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	}

	switch resp.StatusCode {
	case 200: // Headers are set and returned
		return nil
	case 404:
		return usecases.ErrYamsObjectNotFound
	case 500: // Server error
		return usecases.ErrYamsInternal
	case 503: // Service temporarily unavailable
		return usecases.ErrYamsInternal
	default: // Unkown error
		return usecases.ErrYamsInternal
	}
}

func stringConcat(args ...string) string {
	sb := strings.Builder{}
	for _, arg := range args {
		sb.WriteString(arg)
	}
	return sb.String()
}

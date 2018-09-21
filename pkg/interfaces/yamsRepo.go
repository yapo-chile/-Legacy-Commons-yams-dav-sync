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
}

func NewYamsRepository(jwtSigner infra.JWTSigner, mgmtURL, accessKeyID, tenantID, domainID string) *YamsRepository {
	yamsRepo := &YamsRepository{
		jwtSigner:   jwtSigner,
		mgmtURL:     mgmtURL,
		accessKeyID: accessKeyID,
		tenantID:    tenantID,
		domainID:    domainID,
	}

	// Disable debug by default
	yamsRepo.Debug = false

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

func (repo *YamsRepository) PutImage(bucketID string, image domain.Image) *usecases.YamsRepositoryError {

	type Metadata struct {
		ObjectId string `json:"oid"`
	}

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
		stringConcat("POST\\/tenants/", repo.tenantID, "/domains/", repo.domainID, "/buckets/", bucketID, "/objects"),
		Metadata{
			ObjectId: image.Metadata.ImageName,
		},
	}

	tokenString := repo.jwtSigner.GenerateTokenString(claims)

	requestURI := stringConcat("https://", repo.mgmtURL, "/api/v1/tenants/", repo.tenantID, "/domains/", repo.domainID,
		"/buckets/", bucketID, "/objects?jwt=", tokenString, "&AccessKeyId=", repo.accessKeyID)

	if repo.Debug {
		fmt.Println(requestURI)
	}

	imageFile, err := os.Open(image.FilePath)
	if err != nil {
		return usecases.ErrYamsImage
	}
	defer imageFile.Close()

	// TODO: Use connection pull with keepalive
	resp, err := http.Post(requestURI, "images/jpg", imageFile)
	if err != nil {
		return usecases.ErrYamsConnection
	}
	defer resp.Body.Close()

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

	body, _ := ioutil.ReadAll(resp.Body)
	if repo.Debug {
		fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	}

	return nil
}

func stringConcat(args ...string) string {
	sb := strings.Builder{}
	for _, arg := range args {
		sb.WriteString(arg)
	}
	return sb.String()
}

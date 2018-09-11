package interfaces

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/domain"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
)

type YamsRepo struct {
	JWTSigner   infra.JWTSigner
	MgmtURL     string
	AccessKeyID string
	TenantID    string
	DomainID    string
}

func (repo YamsRepo) GetDomains() string {

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
		"GET\\/tenants/" + repo.TenantID + "/domains",
		Metadata{},
	}

	tokenString := repo.JWTSigner.GenerateTokenString(claims)

	requestURI := "https://" + repo.MgmtURL + "/api/v1/tenants/" + repo.TenantID + "/domains?jwt=" + tokenString + "&AccessKeyId=" + repo.AccessKeyID
	fmt.Println(requestURI)

	resp, err := http.Get(requestURI)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)

	return ""
}

func (repo YamsRepo) PutImage(bucketID string, image domain.Image) error {

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
		stringConcat("POST\\/tenants/", repo.TenantID, "/domains/", repo.DomainID, "/buckets/", bucketID, "/objects"),
		Metadata{
			ObjectId: image.Metadata.ImageName,
		},
	}

	tokenString := repo.JWTSigner.GenerateTokenString(claims)

	requestURI := stringConcat("https://", repo.MgmtURL, "/api/v1/tenants/", repo.TenantID, "/domains/", repo.DomainID,
		"/buckets/", bucketID, "/objects?jwt=", tokenString, "&AccessKeyId=", repo.AccessKeyID)
	fmt.Println(requestURI)

	imageFile, err := os.Open(image.FilePath)
	if err != nil {
		return errors.New("Unable to open image file. Detail: " + err.Error())
	}
	defer imageFile.Close()

	resp, err := http.Post(requestURI, "images/jpg", imageFile)
	if err != nil {
		return errors.New("Failed post resquest to Yams. Detail: " + err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
	return nil
}

func stringConcat(args ...string) string {
	sb := strings.Builder{}
	for _, arg := range args {
		sb.WriteString(arg)
	}
	return sb.String()
}

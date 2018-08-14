package interfaces

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	infra "github.schibsted.io/Yapo/yams-dav-sync/pkg/infrastructure"
)

type YamsRepo struct {
	MgmtURL     string
	AccessKeyID string
	TenantID    string
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

	tokenString := infra.GenerateTokenString(claims)

	requestURI := "https://" + repo.MgmtURL + "/api/v1/tenants/" + repo.TenantID + "/domains?jwt=" + tokenString + "&AccessKeyId=" + repo.AccessKeyID
	fmt.Println(requestURI)

	resp, err := http.Get(requestURI)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)

	return ""
}

const MgmtYamsHost = "mgmt-us-east-1-yams.schibsted.com"
const AccessKeyID = "17c82c157c50a0c4"
const TenantID = "e5ce1008-0145-4b91-9670-390db782ed9c"

func putImage(filename string) {

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
		"POST\\/tenants/" + TenantID + "/domains/fa5881b0-3092-4c80-b37b-0ab08519951f/buckets/465a8c49-cdb8-4fe4-8ce2-204860780391/objects",
		Metadata{
			ObjectId: filename,
		},
	}

	tokenString := infra.GenerateTokenString(claims)

	requestURI := "https://" + MgmtYamsHost + "/api/v1/tenants/" + TenantID + "/domains/fa5881b0-3092-4c80-b37b-0ab08519951f/buckets/465a8c49-cdb8-4fe4-8ce2-204860780391/objects?jwt=" + tokenString + "&AccessKeyId=" + AccessKeyID
	fmt.Println(requestURI)

	image, _ := os.Open(filename)
	defer image.Close()

	resp, err := http.Post(requestURI, "image/jpg", image)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("  Response: ", resp, "\n  Body: ", string(body), "\n  Error: ", err)
}

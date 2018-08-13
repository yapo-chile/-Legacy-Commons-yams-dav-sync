package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func main() {
	filename := os.Args[1]
	fmt.Printf("Filename %v\n", filename)
	putImage(filename)
}

func getTenants() {
	accessKeyId := "17c82c157c50a0c4"
	mySigningKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDaLAu3fOIwAfev
+p7YZNm7hhb77hxFsUUrrL3CGZN6HbpHwlWdLZQ3MXbu1eaw+KHKy9wXQlKSikHU
XqOyJmTgQLa86oQeZh+9xKh/lkKlOhbr4OWPuP/WVNwHHl40L6MqOcgQi2XWK+rk
zOVJibZ9HJQYYeX/I5ZWOQq9XxHcgciK5OVnZif+MVn06w7FUkGFpXK2CYjaNssS
G1WxritFWKaWulMKMhXSXJdE6n0dLGec5cC8qgcTivSxAiyw1xMyIQDp7lt2gYYa
5jEFgIT9QsGYR/ALgZOVtGvSYDt4saUuLnzjeRgFNvtc1E7TElDVJjlPhqwAEIG5
yjBOWmPXAgMBAAECggEBAILh8ZV+IeNTCthTrX974PMSmi9AxU0575kn5B7RYRvd
QadS+jF06nnB/uuA/wgj4EvcyIhbjVXEA4H31JRHXDT8HaRvvBrQgTJDDCKebhTZ
KG8RWCZEgZZ/ChBrM3BzM5tdVpw2RD2b0yq3kMXxC706q82EzYmmO8Y2Ki85HWn+
hgnHVj80eCI/OitJvMvOiypYLiPTRhlk+5u6/7Us0QLq1kkzzBMrrwsW7nrscRWT
PqjUUT9s9PxWX0sYz2l081VEzM2DmMtvkYWvBRxlh87pAYaXqju1UGraUcCtUBkr
e+v9QwCAOx6PTiE2ihgSVcUyNjsaNhkCGzMROK8BKUECgYEA92rCiugmYPUouWKl
Er86KkfLS9X/D/PUPoYkhoU/2tE79pWLcpppk5WXZ0cKGmE4mt2iE9IxQ8XTsYeR
UEI50RMEC1a7G7oTw0pAPD2KOXZCMzyEaZDHDbJGjk06nqzkcajFX4O3e2OEyUqZ
6HlmWkMLSqpDJNYp9+qOfEXiw7kCgYEA4b2RzS8xqkogELuA+gjH9LMc2dU5Dg61
jJFxVJOr3w3BNVSlHUkOKmuc7+hyvA6DmEPk4f+AhZFOy/f64p26XzMkImJPd5Vp
EQwUjX0rKnJlHUxPpODhqxssq3epWE5//uMJOOR9EBlBLiKN4Tch/SMgP8tdg3G7
e+TFHsqmTA8CgYEAifwKf3m1XcGccrenJGttvwLHSIYSeA0eQ7iASl2qHRkv/fet
C78+PkbhZ4HhCpFKBmSw7aj+PLPpukrHKiGlKQsX6FL4iyAdwX55kJ8ppZ5kkTqh
BlbuDJ9uZhKALNpzlUfwu7Iz9CaueayXZWW/RXu6omuOgs7GHTO9P2WicFECgYEA
tk5BdKKbinoooTq8c7Epwu3v6+AuHrM0XVyMWRCVaiMSmP5OBnOcdBfKe1mPZ16V
Wh/itb2BTt1F9KXFQMS+4elMUlRw9xN78Z9+7bFbbgFKtbmOTIqs1WGx1pxh8AYd
inxSU1b7xUeQAzE2wd6jnWqDveGAGQp9rhXYOADTAnMCgYEAqKArkBtuvz8kz14J
+uIqUokJvzCnuC5rCC5MEHRVb3xLjdhM94ywyrOzvLclzrCWgrpP4m8188j2muau
IWhfPMvLHyd0+br6IbLKaVL7FBU2F+lHp9R4FwwOiVdjleFJ03H6XmNiyfOJwbqB
RnP+ovKRPliceu47SsDcckezhqY=
-----END RSA PRIVATE KEY-----`)

	block, _ := pem.Decode(mySigningKey)

	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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
		"GET\\/tenants/e5ce1008-0145-4b91-9670-390db782ed9c/domains",
		Metadata{},
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(rsaKey)

	requestUri := "https://mgmt-yams.schibsted.com/api/v1/tenants/e5ce1008-0145-4b91-9670-390db782ed9c/domains?jwt=" + tokenString + "&AccessKeyId=" + accessKeyId
	fmt.Println(requestUri)

	resp, err := http.Get(requestUri)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Response: ", resp, "; Body: ", string(body), "; Error: ", err)
}

func putImage(filename string) {
	accessKeyId := "17c82c157c50a0c4"
	mySigningKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDaLAu3fOIwAfev
+p7YZNm7hhb77hxFsUUrrL3CGZN6HbpHwlWdLZQ3MXbu1eaw+KHKy9wXQlKSikHU
XqOyJmTgQLa86oQeZh+9xKh/lkKlOhbr4OWPuP/WVNwHHl40L6MqOcgQi2XWK+rk
zOVJibZ9HJQYYeX/I5ZWOQq9XxHcgciK5OVnZif+MVn06w7FUkGFpXK2CYjaNssS
G1WxritFWKaWulMKMhXSXJdE6n0dLGec5cC8qgcTivSxAiyw1xMyIQDp7lt2gYYa
5jEFgIT9QsGYR/ALgZOVtGvSYDt4saUuLnzjeRgFNvtc1E7TElDVJjlPhqwAEIG5
yjBOWmPXAgMBAAECggEBAILh8ZV+IeNTCthTrX974PMSmi9AxU0575kn5B7RYRvd
QadS+jF06nnB/uuA/wgj4EvcyIhbjVXEA4H31JRHXDT8HaRvvBrQgTJDDCKebhTZ
KG8RWCZEgZZ/ChBrM3BzM5tdVpw2RD2b0yq3kMXxC706q82EzYmmO8Y2Ki85HWn+
hgnHVj80eCI/OitJvMvOiypYLiPTRhlk+5u6/7Us0QLq1kkzzBMrrwsW7nrscRWT
PqjUUT9s9PxWX0sYz2l081VEzM2DmMtvkYWvBRxlh87pAYaXqju1UGraUcCtUBkr
e+v9QwCAOx6PTiE2ihgSVcUyNjsaNhkCGzMROK8BKUECgYEA92rCiugmYPUouWKl
Er86KkfLS9X/D/PUPoYkhoU/2tE79pWLcpppk5WXZ0cKGmE4mt2iE9IxQ8XTsYeR
UEI50RMEC1a7G7oTw0pAPD2KOXZCMzyEaZDHDbJGjk06nqzkcajFX4O3e2OEyUqZ
6HlmWkMLSqpDJNYp9+qOfEXiw7kCgYEA4b2RzS8xqkogELuA+gjH9LMc2dU5Dg61
jJFxVJOr3w3BNVSlHUkOKmuc7+hyvA6DmEPk4f+AhZFOy/f64p26XzMkImJPd5Vp
EQwUjX0rKnJlHUxPpODhqxssq3epWE5//uMJOOR9EBlBLiKN4Tch/SMgP8tdg3G7
e+TFHsqmTA8CgYEAifwKf3m1XcGccrenJGttvwLHSIYSeA0eQ7iASl2qHRkv/fet
C78+PkbhZ4HhCpFKBmSw7aj+PLPpukrHKiGlKQsX6FL4iyAdwX55kJ8ppZ5kkTqh
BlbuDJ9uZhKALNpzlUfwu7Iz9CaueayXZWW/RXu6omuOgs7GHTO9P2WicFECgYEA
tk5BdKKbinoooTq8c7Epwu3v6+AuHrM0XVyMWRCVaiMSmP5OBnOcdBfKe1mPZ16V
Wh/itb2BTt1F9KXFQMS+4elMUlRw9xN78Z9+7bFbbgFKtbmOTIqs1WGx1pxh8AYd
inxSU1b7xUeQAzE2wd6jnWqDveGAGQp9rhXYOADTAnMCgYEAqKArkBtuvz8kz14J
+uIqUokJvzCnuC5rCC5MEHRVb3xLjdhM94ywyrOzvLclzrCWgrpP4m8188j2muau
IWhfPMvLHyd0+br6IbLKaVL7FBU2F+lHp9R4FwwOiVdjleFJ03H6XmNiyfOJwbqB
RnP+ovKRPliceu47SsDcckezhqY=
-----END RSA PRIVATE KEY-----`)

	block, _ := pem.Decode(mySigningKey)

	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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
		"POST\\/tenants/e5ce1008-0145-4b91-9670-390db782ed9c/domains/fa5881b0-3092-4c80-b37b-0ab08519951f/buckets/465a8c49-cdb8-4fe4-8ce2-204860780391/objects",
		Metadata{
			ObjectId: filename,
		},
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(rsaKey)

	requestUri := "https://mgmt-yams.schibsted.com/api/v1/tenants/e5ce1008-0145-4b91-9670-390db782ed9c/domains/fa5881b0-3092-4c80-b37b-0ab08519951f/buckets/465a8c49-cdb8-4fe4-8ce2-204860780391/objects?jwt=" + tokenString + "&AccessKeyId=" + accessKeyId
	fmt.Println(requestUri)

	image, _ := os.Open(filename)
	defer image.Close()

	resp, err := http.Post(requestUri, "image/jpg", image)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("Response: ", resp, "; Body: ", string(body), "; Error: ", err)
}

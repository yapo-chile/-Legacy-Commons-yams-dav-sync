package infrastructure

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// jwtSSigner generates a digital JWT signature for each request sent to yams
type jwtSigner struct {
	privateKeyFile string
}

// NewJWTSigner returns a new instance of JWTSigner
func NewJWTSigner(privateKetyFile string) repository.Signer {
	return &jwtSigner{
		privateKeyFile: privateKetyFile,
	}
}

func (signer *jwtSigner) getRSAKey() interface{} {
	mySigningKey, _ := ioutil.ReadFile(signer.privateKeyFile)
	block, _ := pem.Decode(mySigningKey)

	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return rsaKey
}

// GenerateTokenString Create a new token object, specifying signing method and the claims
func (signer *jwtSigner) GenerateTokenString(claims jwt.Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(signer.getRSAKey())
	return tokenString
}

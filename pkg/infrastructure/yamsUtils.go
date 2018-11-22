package infrastructure

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

type JWTSigner struct {
	privateKeyFile string
}

func NewJWTSigner(privateKetyFile string) repository.Signer {
	signer := JWTSigner{
		privateKeyFile: privateKetyFile,
	}
	return &signer
}

func (signer *JWTSigner) getRSAKey() interface{} {
	mySigningKey, _ := ioutil.ReadFile(signer.privateKeyFile)
	block, _ := pem.Decode(mySigningKey)

	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return rsaKey
}

func (signer *JWTSigner) GenerateTokenString(claims jwt.Claims) string {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(signer.getRSAKey())

	return tokenString
}

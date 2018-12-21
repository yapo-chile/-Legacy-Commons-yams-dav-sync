package infrastructure

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/loggers"
	"github.schibsted.io/Yapo/yams-dav-sync/pkg/interfaces/repository"
)

// jwtSSigner generates a digital JWT signature for each request sent to yams
type jwtSigner struct {
	privateKeyFile string
	logger         loggers.Logger
}

// NewJWTSigner returns a new instance of JWTSigner
func NewJWTSigner(privateKetyFile string, logger loggers.Logger) repository.Signer {
	return &jwtSigner{
		privateKeyFile: privateKetyFile,
		logger:         logger,
	}
}

func (signer *jwtSigner) getRSAKey() interface{} {
	mySigningKey, err := ioutil.ReadFile(signer.privateKeyFile)
	if err != nil {
		signer.logger.Error("Error reading Private Key: %+v", err)
		return nil
	}
	block, _ := pem.Decode(mySigningKey)

	rsaKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		signer.logger.Error("Error parsing Private Key: %+v", err)
		return nil
	}
	return rsaKey
}

// GenerateTokenString Create a new token object, specifying signing method and the claims
func (signer *jwtSigner) GenerateTokenString(claims jwt.Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(signer.getRSAKey())
	if err != nil {
		signer.logger.Error("Error with signature for claims: %+v", err)
	}
	return tokenString
}

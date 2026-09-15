// auth package
//
//

package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
)

func HashPassword(password string) (string, error) {

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		fmt.Printf("error hashing password: %s", err)
	}

	return hash, err

}

func CheckPasswrodHash(password, hash string) (bool, error) {

	matched, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		fmt.Printf("error comparing password and hash: %s", err)
	}
	return matched, err

}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	var claims jwt.RegisteredClaims
	claims.Issuer = string(TokenTypeAccess)
	claims.IssuedAt = jwt.NewNumericDate(time.Now().UTC())
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(expiresIn))
	claims.Subject = userID.String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	var claims jwt.RegisteredClaims

	token, err := jwt.ParseWithClaims(tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil })
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	subjectUUID, err := uuid.Parse(subject)
	if err != nil {
		log.Printf("error parsing subject UUID: %s", err)
	}

	return subjectUUID, err
}

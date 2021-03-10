package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//Claims ...
type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

//GeneratePasswordResetToken using jwt-go...
func GeneratePasswordResetToken(currentPassword string, email string, userID int64) (string, error) {
	secret := []byte(os.Getenv("PASSWORD_RESET_SECRET") + currentPassword)

	//tokenClaims
	claims := Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			Subject:   fmt.Sprint(userID),
			IssuedAt:  time.Now().Local().Unix(),
			ExpiresAt: time.Now().Local().Add(time.Minute * 10).Unix(),
			Issuer:    os.Getenv("TITLE"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	passwordResetToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return passwordResetToken, nil
}

//ValidatePasswordResetToken ...
func ValidatePasswordResetToken(currentPassword string, passwordResetToken string) (Claims, error) {
	secret := []byte(os.Getenv("PASSWORD_RESET_SECRET") + currentPassword)

	var claims Claims
	token, err := jwt.ParseWithClaims(
		passwordResetToken,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
	)
	log.Println(claims)
	if err == nil && token.Valid {
		return claims, nil
	}
	return Claims{}, err
}

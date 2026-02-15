package helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

func HashRequestPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(user model.User) (string, error) {
	claims := model.CustomClaims{
		UserID: user.ID,
		RoleID: user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func DecodeToken(token string, claim *model.CustomClaims) (err error) {
	jwt.ParseWithClaims(token, claim, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSigningKey()), nil
	})
	return
}

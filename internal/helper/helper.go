package helper

import (
	"context"
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
		Role:   user.Role.Name,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func DecodeToken(tokenString string, claim *model.CustomClaims) error {
	token, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		return []byte(config.JWTSigningKey()), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}

	return nil
}

func GetUserFromContext(ctx context.Context) *model.CustomClaims {
	claims, ok := ctx.Value(model.BearerAuthKey).(*model.CustomClaims)
	if !ok {
		return nil
	}
	return claims
}

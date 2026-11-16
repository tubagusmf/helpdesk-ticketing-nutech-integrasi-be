package helper

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

func GenerateTokenWithTTL(user model.User, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 10 * time.Hour
	}

	signingKey := strings.TrimSpace(
		config.JWTSigningKey(),
	)

	if signingKey == "" {
		return "", errors.New(
			"JWT_SIGNING_KEY belum dikonfigurasi",
		)
	}

	now := time.Now()

	claims := model.CustomClaims{
		UserID: user.ID,
		RoleID: user.RoleID,
		Role:   user.Role.Name,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(
				now.Add(ttl),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(
		[]byte(signingKey),
	)
}

func GenerateToken(user model.User) (string, error) {
	return GenerateTokenWithTTL(
		user,
		10*time.Hour,
	)
}

func DecodeToken(tokenString string, claim *model.CustomClaims) error {
	signingKey := strings.TrimSpace(config.JWTSigningKey())

	if signingKey == "" {
		return errors.New("JWT signing key is not configured")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claim,
		func(t *jwt.Token) (interface{}, error) {

			if t.Method == nil ||
				t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New(
					"unexpected JWT signing algorithm",
				)
			}

			return []byte(signingKey), nil
		},
	)

	if err != nil {
		return err
	}

	if token == nil || !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}

	if claim.ExpiresAt == nil {
		return errors.New("JWT expiration is required")
	}

	if !claim.VerifyExpiresAt(time.Now(), true) {
		return jwt.ErrTokenExpired
	}

	if claim.UserID <= 0 || claim.RoleID <= 0 {
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

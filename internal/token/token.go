package token

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtCfg *JWT

	ctxKey = &struct{}{}
)

type JWT struct {
	secret string
	ttl    time.Duration
}

func init() {
	jwtCfg = &JWT{}
	var err error
	jwtCfg.secret = os.Getenv("JWT_SECRET")
	jwtCfg.ttl, err = time.ParseDuration(os.Getenv("JWT_TTL"))
	if err != nil {
		jwtCfg.ttl = 10 * time.Minute
	}
}

func Create(user string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "avito-merch-shop",
		Subject:   user,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtCfg.ttl)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtCfg.secret))
}

func Check(token string) (string, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) { return []byte(jwtCfg.secret), nil }, jwt.WithExpirationRequired())
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, ctxKey, user)
}

func UserFromContext(ctx context.Context) string {
	v, ok := ctx.Value(ctxKey).(string)
	if !ok {
		return ""
	}
	return v
}

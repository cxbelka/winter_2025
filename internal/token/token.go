package token

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTokenTTL = 10 * time.Minute
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
	Reinit()
}

func Reinit() {
	jwtCfg = &JWT{}
	var err error
	jwtCfg.secret = os.Getenv("JWT_SECRET")
	jwtCfg.ttl, err = time.ParseDuration(os.Getenv("JWT_TTL"))
	if err != nil {
		jwtCfg.ttl = defaultTokenTTL
	}
}

func Create(user string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "avito-merch-shop",
		Subject:   user,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtCfg.ttl)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtCfg.secret)) //nolint:wrapcheck
}

func Check(token string) (string, error) {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(
		token, &claims,
		func(_ *jwt.Token) (interface{}, error) { return []byte(jwtCfg.secret), nil },
		jwt.WithExpirationRequired(),
	)

	return claims.Subject, err //nolint:wrapcheck
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

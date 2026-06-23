package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret []byte
	TTL    time.Duration
	Issuer string
}

type Service struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewService(cfg Config) (*Service, error) {
	if len(cfg.Secret) == 0 {
		return nil, errors.New("auth: empty secret")
	}
	return &Service{
		secret: cfg.Secret,
		ttl:    cfg.TTL,
		issuer: cfg.Issuer,
	}, nil
}

type claims struct {
	UserName    string `json:"username"`
	RegistryURL string `json:"registry_url"`
	SessionID   string `json:"sid"`
	jwt.RegisteredClaims
}

// Identity is the non-secret payload carried by the JWT. SessionID references
// the server-side session that holds the actual registry credentials.
type Identity struct {
	UserName    string
	RegistryURL string
	SessionID   string
}

func (s *Service) Generate(id Identity) (string, error) {
	now := time.Now()
	claims := claims{
		UserName:    id.UserName,
		RegistryURL: id.RegistryURL,
		SessionID:   id.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) Decode(tokenString string) (*Identity, error) {
	c := &claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		c,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return &Identity{
		UserName:    c.UserName,
		RegistryURL: c.RegistryURL,
		SessionID:   c.SessionID,
	}, nil
}

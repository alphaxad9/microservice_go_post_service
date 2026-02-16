package auth

import (
	"crypto/rsa"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Verifier struct {
	publicKey *rsa.PublicKey
	lastFetch time.Time
	ttl       time.Duration
	url       string
	mu        sync.Mutex
}

func NewVerifier(url string, ttl time.Duration) *Verifier {
	return &Verifier{
		url: url,
		ttl: ttl,
	}
}

func (v *Verifier) fetchKey() (*rsa.PublicKey, error) {
	resp, err := http.Get(v.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(data)
}

func (v *Verifier) getKey() (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.publicKey != nil && time.Since(v.lastFetch) < v.ttl {
		return v.publicKey, nil
	}

	key, err := v.fetchKey()
	if err != nil {
		return nil, err
	}

	v.publicKey = key
	v.lastFetch = time.Now()
	return key, nil
}

func (v *Verifier) Verify(token string) (jwt.MapClaims, error) {
	key, err := v.getKey()
	if err != nil {
		return nil, err
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return key, nil
	})

	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if _, ok := claims["user_id"]; !ok {
		return nil, errors.New("user_id missing")
	}

	return claims, nil
}

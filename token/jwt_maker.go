package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMarker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at lease %d charactes", minSecretKeySize)
	}
	return &JWTMarker{secretKey}, nil
}

func (maker *JWTMarker) CreateToken(username string, duration time.Duration) (string, *PasetoPayload, error) {
	payload, err := NewPayload(username, duration)
	pasetoPayload, err1 := NewPasetoPayload(username, duration)
	if err != nil || err1 != nil {
		return "", nil, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, pasetoPayload, err
}

func (maker *JWTMarker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		errString := err.Error()
		if strings.Contains(errString, "expired") {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)

	if !ok {
		return nil, ErrInvalidToken
	}
	if time.Now().After(payload.RegisteredClaims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	return payload, nil
}

func (maker *JWTMarker) VerifyPasetoToken(token string) (*PasetoPayload, error) {
	return nil, nil
}

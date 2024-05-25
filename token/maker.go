package token

import (
	"time"
)

type Maker interface {
	CreateToken(username string, duration time.Duration) (string, *PasetoPayload, error)
	VerifyToken(token string) (*Payload, error)
	VerifyPasetoToken(token string) (*PasetoPayload, error)
}

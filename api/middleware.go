package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/abhilashdk2016/golang-simple-bank/token"
	"github.com/gin-gonic/gin"
)

const (
	autorizationHeaderKey  = "authorization"
	autorizationTypeBearer = "bearer"
	autorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(autorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("autorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid autorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != autorizationTypeBearer {
			err := errors.New("autorization type not supported by server")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyPasetoToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(autorizationPayloadKey, payload)
		ctx.Next()
	}
}

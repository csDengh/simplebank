package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/csdengh/cur_blank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func AuthenticateMideware(tm token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		token := fields[1]
		pl, err := tm.ValidToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorRes(err))
			return
		}

		ctx.Set(authorizationPayloadKey, pl)
		ctx.Next()
	}
}

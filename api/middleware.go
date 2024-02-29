package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	db "github.com/40grivenprog/simple-bank/db/sqlc"
	"github.com/40grivenprog/simple-bank/jaeger"
	"github.com/40grivenprog/simple-bank/token"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		spanCtx, _ := jaeger.Extract(jaeger.Tracer, c.Request)
		span := jaeger.Tracer.StartSpan(c.Request.URL.Path, ext.RPCServerOption(spanCtx))
		defer span.Finish()

		err := jaeger.Inject(span, c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to inject span context"})
			return
		}

		c.Set("span", span)

		c.Next()
	}
}

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("auth header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("auth header is not valid")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("auth header type is not valid %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func roleMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
		if authPayload.Role != db.UserRoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errorResponse(errors.New("Not authorized to perform this action")))
			return
		}
	}
}

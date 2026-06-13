package auth

import (
	"net/http"
	"strings"

	"stock-market/backend/internal/model"

	"github.com/gin-gonic/gin"
)

const authUserKey = "authUser"

func Authenticate(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := getBearerToken(c)
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.AuthErrorResponse{
				Error:     "Access token missing",
				ErrorCode: "ACCESS_TOKEN_MISSING",
			})
			return
		}

		payload, err := VerifyAccessToken(cfg, accessToken)
		if err != nil {
			if IsTokenExpired(err) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.AuthErrorResponse{
					Error:     "Access token expired",
					ErrorCode: "ACCESS_TOKEN_EXPIRED",
				})
				return
			}
			if IsTokenInvalid(err) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.AuthErrorResponse{
					Error:     "Invalid access token",
					ErrorCode: "ACCESS_TOKEN_INVALID",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, model.AuthErrorResponse{
				Error:     "Unauthorized",
				ErrorCode: "UNAUTHORIZED",
			})
			return
		}

		c.Set(authUserKey, payload)
		c.Next()
	}
}

func GetAuthUser(c *gin.Context) (model.AuthTokenPayload, bool) {
	value, ok := c.Get(authUserKey)
	if !ok {
		return model.AuthTokenPayload{}, false
	}

	payload, ok := value.(model.AuthTokenPayload)
	return payload, ok
}

func getBearerToken(c *gin.Context) string {
	authorizationHeader := c.GetHeader("Authorization")
	if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
		if token != "" {
			return token
		}
	}

	return strings.TrimSpace(c.Query("access_token"))
}

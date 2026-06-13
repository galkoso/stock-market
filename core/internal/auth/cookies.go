package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetRefreshTokenCookie(c *gin.Context, cfg Config, refreshToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		cfg.RefreshTokenCookie,
		refreshToken,
		cfg.RefreshTokenExpiresInSeconds,
		"/",
		"",
		cfg.SecureCookies,
		true,
	)
}

func ClearRefreshTokenCookie(c *gin.Context, cfg Config) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		cfg.RefreshTokenCookie,
		"",
		-1,
		"/",
		"",
		cfg.SecureCookies,
		true,
	)
}

func GetRefreshTokenFromCookie(c *gin.Context, cfg Config) (string, error) {
	return c.Cookie(cfg.RefreshTokenCookie)
}

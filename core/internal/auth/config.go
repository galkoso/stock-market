package auth

import "time"

type Config struct {
	RefreshTokenCookie            string
	AccessJWTSecret               string
	RefreshJWTSecret              string
	AccessTokenExpiresInSeconds   int
	RefreshTokenExpiresInSeconds  int
	SecureCookies                 bool
}

func (c Config) AccessTokenTTL() time.Duration {
	return time.Duration(c.AccessTokenExpiresInSeconds) * time.Second
}

func (c Config) RefreshTokenTTL() time.Duration {
	return time.Duration(c.RefreshTokenExpiresInSeconds) * time.Second
}

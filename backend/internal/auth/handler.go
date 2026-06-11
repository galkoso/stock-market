package auth

import (
	"errors"
	"net/http"

	"stock-market/backend/internal/model"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	cfg     Config
}

func NewHandler(service *Service, cfg Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

func (h *Handler) Register(c *gin.Context) {
	var body model.RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.AuthErrorResponse{Error: "Username and password are required"})
		return
	}

	session, err := h.service.Register(c.Request.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingFields):
			c.JSON(http.StatusBadRequest, model.AuthErrorResponse{Error: err.Error()})
		case errors.Is(err, ErrUserExists):
			c.JSON(http.StatusBadRequest, model.AuthErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, model.AuthErrorResponse{Error: err.Error()})
		}
		return
	}

	SetRefreshTokenCookie(c, h.cfg, session.RefreshToken)
	c.JSON(http.StatusOK, model.AuthSuccessResponse{
		Success:     true,
		UserID:      session.User.ID,
		User:        &session.User,
		AccessToken: session.AccessToken,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var body model.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.AuthErrorResponse{Error: "Username and password are required"})
		return
	}

	session, err := h.service.Login(c.Request.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingFields):
			c.JSON(http.StatusBadRequest, model.AuthErrorResponse{Error: err.Error()})
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, model.AuthErrorResponse{Error: err.Error()})
		}
		return
	}

	SetRefreshTokenCookie(c, h.cfg, session.RefreshToken)
	c.JSON(http.StatusOK, model.AuthSuccessResponse{
		Success:     true,
		User:        &session.User,
		AccessToken: session.AccessToken,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := GetRefreshTokenFromCookie(c, h.cfg)
	if err != nil || refreshToken == "" {
		ClearRefreshTokenCookie(c, h.cfg)
		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{
			Error:     "Refresh token missing",
			ErrorCode: "REFRESH_TOKEN_MISSING",
		})
		return
	}

	session, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		ClearRefreshTokenCookie(c, h.cfg)
		if IsTokenExpired(err) {
			c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{
				Error:     "Refresh token expired",
				ErrorCode: "REFRESH_TOKEN_INVALID",
			})
			return
		}
		if IsTokenInvalid(err) || errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{
				Error:     "Invalid refresh token",
				ErrorCode: "REFRESH_TOKEN_INVALID",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, model.AuthErrorResponse{
			Error:     "Refresh token invalid or expired",
			ErrorCode: "REFRESH_TOKEN_INVALID",
		})
		return
	}

	SetRefreshTokenCookie(c, h.cfg, session.RefreshToken)
	c.JSON(http.StatusOK, model.AuthSuccessResponse{
		Success:     true,
		User:        &session.User,
		AccessToken: session.AccessToken,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	ClearRefreshTokenCookie(c, h.cfg)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

package model

type User struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Password string `json:"-" bson:"password"`
}

type SafeUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthTokenSubject struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

type AuthTokenPayload struct {
	AuthTokenSubject
	TokenType string `json:"tokenType"`
}

type AuthSession struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	User         SafeUser `json:"user"`
}

type AuthSuccessResponse struct {
	Success     bool      `json:"success"`
	User        *SafeUser `json:"user,omitempty"`
	AccessToken string    `json:"accessToken,omitempty"`
	UserID      string    `json:"userId,omitempty"`
}

type AuthErrorResponse struct {
	Error     string `json:"error"`
	ErrorCode string `json:"errorCode,omitempty"`
}

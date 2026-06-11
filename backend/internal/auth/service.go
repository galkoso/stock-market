package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"stock-market/backend/internal/model"
	mongopkg "stock-market/backend/internal/mongo"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrMissingFields      = errors.New("username and password are required")
)

type Service struct {
	cfg   Config
	users *mongo.Collection
}

func NewService(cfg Config, users *mongo.Collection) *Service {
	return &Service{cfg: cfg, users: users}
}

func (s *Service) Register(ctx context.Context, input model.RegisterRequest) (model.AuthSession, error) {
	username := normalizeUsername(input.Username)
	password := strings.TrimSpace(input.Password)

	if username == "" || password == "" {
		return model.AuthSession{}, ErrMissingFields
	}

	if err := s.users.FindOne(ctx, bson.M{"username": username}).Err(); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return model.AuthSession{}, fmt.Errorf("lookup user: %w", err)
		}
	} else {
		return model.AuthSession{}, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.AuthSession{}, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		ID:       uuid.NewString(),
		Username: username,
		Password: string(hashedPassword),
	}

	if _, err := s.users.InsertOne(ctx, bson.M{
		"_id":      user.ID,
		"username": user.Username,
		"password": user.Password,
	}); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.AuthSession{}, ErrUserExists
		}
		return model.AuthSession{}, fmt.Errorf("insert user: %w", err)
	}

	return s.buildSession(user)
}

func (s *Service) Login(ctx context.Context, credentials model.LoginRequest) (model.AuthSession, error) {
	username := normalizeUsername(credentials.Username)
	password := strings.TrimSpace(credentials.Password)

	if username == "" || password == "" {
		return model.AuthSession{}, ErrMissingFields
	}

	var user model.User
	if err := s.users.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.AuthSession{}, ErrInvalidCredentials
		}
		return model.AuthSession{}, fmt.Errorf("lookup user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.AuthSession{}, ErrInvalidCredentials
	}

	return s.buildSession(user)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (model.AuthSession, error) {
	payload, err := VerifyRefreshToken(s.cfg, refreshToken)
	if err != nil {
		return model.AuthSession{}, err
	}

	var user model.User
	if err := s.users.FindOne(ctx, bson.M{"_id": payload.UserID}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.AuthSession{}, ErrUserNotFound
		}
		return model.AuthSession{}, fmt.Errorf("lookup user: %w", err)
	}

	return s.buildSession(user)
}

func (s *Service) buildSession(user model.User) (model.AuthSession, error) {
	subject := model.AuthTokenSubject{
		UserID:   user.ID,
		Username: user.Username,
	}

	accessToken, err := GenerateAccessToken(s.cfg, subject)
	if err != nil {
		return model.AuthSession{}, err
	}

	refreshToken, err := GenerateRefreshToken(s.cfg, subject)
	if err != nil {
		return model.AuthSession{}, err
	}

	return model.AuthSession{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         mongopkg.ToSafeUser(user),
	}, nil
}

func normalizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

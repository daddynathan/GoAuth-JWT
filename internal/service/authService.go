package service

import (
	"context"
	"errors"
	"fmt"
	"friend-help/internal/cache"
	"friend-help/internal/errs"
	"friend-help/internal/model"
	"friend-help/internal/repo"
	"regexp"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo     repo.AuthRepo
	jwtService   *JwtService
	redisService *cache.RedisService
}

func NewAuthService(authRepo repo.AuthRepo, JwtService *JwtService, redisService *cache.RedisService) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		jwtService:   JwtService,
		redisService: redisService,
	}
}

var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateLoginChars(login string) error {
	if !loginRegex.MatchString(login) {
		return fmt.Errorf("%w: Login '%s' must contain only en letters, numbers, underscore", errs.ErrInvalidLoginChars, login)
	}
	return nil
}

func (s *AuthService) RegNewUser(ctx context.Context, req model.AuthRegReq) (int, string, error) {
	if err := ValidateLoginChars(req.Login); err != nil {
		return 0, "", err
	}
	b, err := s.authRepo.CheckUserExists(ctx, req.Login, req.Email)
	if err != nil {
		return 0, "", err
	}
	if b {
		return 0, "", errs.ErrUserExists
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %w", errs.ErrFailedHashPass, err)
	}
	newUser := model.AuthUser{
		Login:        req.Login,
		Email:        &req.Email,
		Username:     req.Login,
		PasswordHash: string(hashedPassword),
		IsActivated:  true,
	}
	userID, err := s.authRepo.CreateUser(ctx, newUser)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %w", errs.ErrFailedToAddUserInDB, err)
	}
	token, err := s.jwtService.GenToken(userID, model.Member)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %w", errs.ErrFailedGenToken, err)
	}
	return userID, token, nil
}

func (s *AuthService) Authenticate(ctx context.Context, identifier string, password string) (*model.AuthUser, string, error) {
	user, err := s.authRepo.GetUserByLoginOrEmail(ctx, identifier)
	if err != nil {
		return nil, "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", errs.ErrInvalidLoginOrPass
		}
		return nil, "", fmt.Errorf("%w: %w", errs.ErrFailedToComparePassHash, err)
	}
	token, err := s.jwtService.GenToken(user.ID, model.Member)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	return user, token, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.jwtService.ParseTokenAndGetClaims(tokenString)
	if err != nil {
		return nil
	}
	if claims.ExpiresAt == nil {
		return errs.ErrInvalidTokenExpTime
	}
	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	if timeUntilExpiry <= 0 {
		return nil
	}
	key := fmt.Sprintf("blacklist:%s", tokenString)
	cmd := s.redisService.Set(ctx, key, claims.UserID, timeUntilExpiry)
	if cmd.Err() != nil {
		return fmt.Errorf("failed to blacklist token in Redis: %w", cmd.Err())
	}
	return nil
}

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenString)
	cmd := s.redisService.Get(ctx, key)
	if cmd.Err() == redis.Nil {
		return false, nil
	}
	if cmd.Err() != nil {
		return false, fmt.Errorf("failed to check blacklist in Redis: %w", cmd.Err())
	}
	return true, nil
}

func (s *AuthService) ParseTokenAndGetClaims(tokenString string) (*model.AuthClaims, error) {
	return s.jwtService.ParseTokenAndGetClaims(tokenString)
}

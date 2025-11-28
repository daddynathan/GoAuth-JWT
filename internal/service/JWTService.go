package service

import (
	"errors"
	"fmt"
	"friend-help/internal/errs"
	"friend-help/internal/model"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	tokenTTL  time.Duration
	secretKey string
}

func NewJwtService() (*JwtService, error) {
	tokenTTLs := os.Getenv("TOKEN_TTL_HOURS")
	if tokenTTLs == "" {
		return nil, errors.New("var TOKEN_TTL_HOURS not found")
	}
	tokenTTLi, err := strconv.Atoi(tokenTTLs)
	if err != nil {
		return nil, fmt.Errorf("var TOKEN_TTL_HOURS bad format: %w", err)
	}
	JWTSecretKey := os.Getenv("JWT_SECRET_KEY")
	if JWTSecretKey == "" {
		return nil, errors.New("var JWT_SECRET_KEY not found")
	}
	return &JwtService{
		tokenTTL:  time.Duration(tokenTTLi) * time.Hour,
		secretKey: JWTSecretKey,
	}, nil
}

func (j *JwtService) GenToken(userID int, role int) (string, error) {
	claims := model.AuthClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JwtService) ParseTokenAndGetClaims(tokenString string) (*model.AuthClaims, error) {
	claims := &model.AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrUnexpectedSigningMethod
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.ErrInvalidToken
	}
	return claims, nil
}

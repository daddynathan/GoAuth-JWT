package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type AuthRegReq struct {
	Login    string `json:"username" binding:"required,min=4,max=32"`
	Email    string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type AuthLogReq struct {
	Identifier string `json:"identifier" binding:"required,max=100"` // Логин ИЛИ Email
	Password   string `json:"password" binding:"required,max=32"`
}

type AuthUser struct {
	ID           int
	Login        string
	Username     string
	Email        *string
	PasswordHash string
	IsActivated  bool
}

type AuthClaims struct {
	UserID int `json:"user_id"`
	Role   int `json:"role"`
	jwt.RegisteredClaims
}

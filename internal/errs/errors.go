package errs

import "errors"

var (
	ErrUserExists         = errors.New("user with this email/login already exists")
	ErrUserNotFound       = errors.New("user with this email/login not found")
	ErrInvalidLoginOrPass = errors.New("invalid login or password")

	ErrFailedHashPass          = errors.New("failed to hash password")
	ErrFailedGenToken          = errors.New("failed to generate token")
	ErrFailedToComparePassHash = errors.New("failed to compare password hash")
	ErrInvalidLoginChars       = errors.New("login contains disallowed characters")

	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("token is invalid")
	ErrInvalidTokenExpTime     = errors.New("token has no expiration time")

	ErrFailedToAddUserInDB = errors.New("failed to create user in DB")
	ErrDBInsertFailed      = errors.New("failed to insert in DB")

	ErrFailedToOpenDB = errors.New("failed to open DB")
	ErrFailedToPingDB = errors.New("failed to ping DB")

	ErrFailedToPingRedis = errors.New("failed to connect to Redis")
)

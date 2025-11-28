package https

import (
	"context"
	"friend-help/internal/model"
)

type contextKey string

const UserCtxKey contextKey = "user_claims"

func GetUserFromContext(ctx context.Context) (*model.AuthClaims, bool) {
	claims, ok := ctx.Value(UserCtxKey).(*model.AuthClaims)
	return claims, ok
}

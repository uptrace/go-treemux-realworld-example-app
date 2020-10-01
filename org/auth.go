package org

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/vmihailenco/treemux"
)

type (
	userCtxKey    struct{}
	userErrCtxKey struct{}
)

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userCtxKey{}).(*User)
	return user
}

func authToken(req treemux.Request) string {
	const prefix = "Token "
	v := req.Header.Get("Authorization")
	v = strings.TrimPrefix(v, prefix)
	return v
}

func UserMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		ctx := req.Context()

		token := authToken(req)
		userID, err := decodeUserToken(token)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		user, err := SelectUser(ctx, userID)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		user.Token, err = CreateUserToken(user.ID, 24*time.Hour)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		ctx = context.WithValue(ctx, userCtxKey{}, user)
		return next(w, req.WithContext(ctx))
	}
}

func MustUserMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		if err, ok := req.Context().Value(userErrCtxKey{}).(error); ok {
			return err
		}
		return next(w, req)
	}
}

package org

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/uptrace/go-realworld-example-app/rwe"
)

type User struct {
	tableName struct{} `pg:",alias:u"`

	ID           uint64 `json:"-"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Bio          string `json:"bio"`
	Image        string `json:"image"`
	Password     string `pg:"-" json:"password,omitempty"`
	PasswordHash string `json:"-"`
	Following    bool   `pg:"-" json:"following"`

	Token string `pg:"-" json:"token,omitempty"`
}

type FollowUser struct {
	tableName struct{} `pg:"alias:fu"`

	UserID         uint64
	FollowedUserID uint64
}

type Profile struct {
	tableName struct{} `pg:"users,alias:u"`

	ID        uint64 `json:"-"`
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `pg:"-" json:"following"`
}

func NewProfile(user *User) *Profile {
	return &Profile{
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Following: user.Following,
	}
}

func SelectUser(ctx context.Context, userID uint64) (*User, error) {
	user := new(User)
	if err := rwe.RedisCache().Once(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("user:%d", userID),
		Value: user,
		TTL:   15 * time.Minute,
		Do: func(item *cache.Item) (interface{}, error) {
			return selectUser(ctx, userID)
		},
	}); err != nil {
		return nil, err
	}
	return user, nil
}

func selectUser(ctx context.Context, id uint64) (*User, error) {
	user := new(User)
	if err := rwe.PGMain().
		ModelContext(ctx, user).
		Where("id = ?", id).
		Select(); err != nil {
		return nil, err
	}
	return user, nil
}

func SelectUserByUsername(ctx context.Context, username string) (*User, error) {
	user := new(User)
	if err := rwe.PGMain().
		ModelContext(ctx, user).
		Where("username = ?", username).
		Select(); err != nil {
		return nil, err
	}

	return user, nil
}

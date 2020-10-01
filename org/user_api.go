package org

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-pg/pg/v10/orm"
	"github.com/vmihailenco/treemux"
	"golang.org/x/crypto/bcrypt"

	"github.com/uptrace/go-realworld-example-app/httputil"
	"github.com/uptrace/go-realworld-example-app/rwe"
)

const kb = 10

var errUserNotFound = errors.New("Not registered email or invalid password")

func setUserToken(user *User) error {
	token, err := CreateUserToken(user.ID, 24*time.Hour)
	if err != nil {
		return err
	}
	user.Token = token
	return nil
}

func currentUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	user := UserFromContext(req.Context())
	return httputil.Write(w, httputil.M{
		"user": user,
	})
}

func createUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	var in struct {
		User *User `json:"user"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := in.User

	var err error
	user.PasswordHash, err = hashPassword(user.Password)
	if err != nil {
		return err
	}

	if _, err := rwe.PGMain().
		ModelContext(ctx, user).
		Insert(); err != nil {
		return err
	}

	if err = setUserToken(user); err != nil {
		return err
	}

	user.Password = ""
	return httputil.Write(w, httputil.M{
		"user": user,
	})
}

func hashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func loginUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	var in struct {
		User *struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := new(User)
	if err := rwe.PGMain().
		ModelContext(ctx, user).
		Where("email = ?", in.User.Email).
		Select(); err != nil {
		return err
	}

	if err := comparePasswords(user.PasswordHash, in.User.Password); err != nil {
		return err
	}

	if err := setUserToken(user); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"user": user,
	})
}

func comparePasswords(hash, pass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return errUserNotFound
	}
	return nil
}

func updateUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	var in struct {
		User *User `json:"user"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := in.User

	var err error
	user.PasswordHash, err = hashPassword(user.Password)
	if err != nil {
		return err
	}

	if _, err = rwe.PGMain().
		ModelContext(ctx, authUser).
		Set("email = ?", user.Email).
		Set("username = ?", user.Username).
		Set("password_hash = ?", user.PasswordHash).
		Set("image = ?", user.Image).
		Set("bio = ?", user.Bio).
		Where("id = ?", authUser.ID).
		Returning("*").
		Update(); err != nil {
		return err
	}

	user.Password = ""
	return httputil.Write(w, httputil.M{
		"user": authUser,
	})
}

func profileEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	followingColumn := func(q *orm.Query) (*orm.Query, error) {
		if authUser, ok := ctx.Value(userCtxKey{}).(*User); ok {
			subq := rwe.PGMain().Model((*FollowUser)(nil)).
				Where("fu.followed_user_id = u.id").
				Where("fu.user_id = ?", authUser.ID)

			q = q.ColumnExpr("EXISTS (?) AS following", subq)
		} else {
			q = q.ColumnExpr("false AS following")
		}

		return q, nil
	}

	user := new(User)
	if err := rwe.PGMain().
		ModelContext(ctx, user).
		ColumnExpr("u.*").
		Apply(followingColumn).
		Where("username = ?", req.Param("username")).
		Select(); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"profile": NewProfile(user),
	})
}

func followUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	user, err := SelectUserByUsername(ctx, req.Param("username"))
	if err != nil {
		return err
	}

	followUser := &FollowUser{
		UserID:         authUser.ID,
		FollowedUserID: user.ID,
	}
	if _, err := rwe.PGMain().
		ModelContext(ctx, followUser).
		Insert(); err != nil {
		return err
	}

	user.Following = true
	return httputil.Write(w, httputil.M{
		"profile": NewProfile(user),
	})
}

func unfollowUserEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	user, err := SelectUserByUsername(ctx, req.Param("username"))
	if err != nil {
		return err
	}

	if _, err := rwe.PGMain().
		ModelContext(ctx, (*FollowUser)(nil)).
		Where("user_id = ?", authUser.ID).
		Where("followed_user_id = ?", user.ID).
		Delete(); err != nil {
		return err
	}

	user.Following = false
	return httputil.Write(w, httputil.M{
		"profile": NewProfile(user),
	})
}

package org

import (
	"github.com/uptrace/go-realworld-example-app/rwe"
)

func init() {
	g := rwe.API.WithMiddleware(UserMiddleware)

	g.POST("/users", createUserHandler)
	g.POST("/users/login", loginUserHandler)
	g.GET("/profiles/:username", profileHandler)

	g = g.WithMiddleware(MustUserMiddleware)

	g.GET("/user/", currentUserHandler)
	g.PUT("/user/", updateUserHandler)

	g.POST("/profiles/:username/follow", followUserHandler)
	g.DELETE("/profiles/:username/follow", unfollowUserHandler)
}

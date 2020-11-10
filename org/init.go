package org

import (
	"github.com/uptrace/go-realworld-example-app/rwe"
)

func init() {
	g := rwe.API.WithMiddleware(UserMiddleware)

	g.POST("/users", createUserEndpoint)
	g.POST("/users/login", loginUserEndpoint)
	g.GET("/profiles/:username", profileEndpoint)

	g = g.WithMiddleware(MustUserMiddleware)

	g.GET("/user/", currentUserEndpoint)
	g.PUT("/user/", updateUserEndpoint)

	g.POST("/profiles/:username/follow", followUserEndpoint)
	g.DELETE("/profiles/:username/follow", unfollowUserEndpoint)
}

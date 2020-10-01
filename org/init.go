package org

import (
	"github.com/uptrace/go-realworld-example-app/rwe"
)

func init() {
	g := rwe.API.NewGroup("")

	g.Use(UserMiddleware)

	g.POST("/users", createUserEndpoint)
	g.POST("/users/login", loginUserEndpoint)
	g.GET("/profiles/:username", profileEndpoint)

	g.Use(MustUserMiddleware)

	g.GET("/user/", currentUserEndpoint)
	g.PUT("/user/", updateUserEndpoint)

	g.POST("/profiles/:username/follow", followUserEndpoint)
	g.DELETE("/profiles/:username/follow", unfollowUserEndpoint)
}

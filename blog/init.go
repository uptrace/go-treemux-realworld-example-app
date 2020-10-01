package blog

import (
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
)

func init() {
	g := rwe.API.NewGroup("")

	g.Use(org.UserMiddleware)

	g.GET("/tags/", listTagsEndpoint)
	g.GET("/articles", listArticlesEndpoint)
	g.GET("/articles/feed", articleFeedEndpoint)
	g.GET("/articles/:slug", showArticleEndpoint)
	g.GET("/articles/:slug/comments", listCommentsEndpoint)
	g.GET("/articles/:slug/comments/:id", showCommentEndpoint)

	g.Use(org.MustUserMiddleware)

	g.POST("/articles", createArticleEndpoint)
	g.PUT("/articles/:slug", updateArticleEndpoint)
	g.DELETE("/articles/:slug", deleteArticleEndpoint)

	g.POST("/articles/:slug/favorite", favoriteArticleEndpoint)
	g.DELETE("/articles/:slug/favorite", unfavoriteArticleEndpoint)

	g.POST("/articles/:slug/comments", createCommentEndpoint)
	g.DELETE("/articles/:slug/comments/:id", deleteCommentEndpoint)
}

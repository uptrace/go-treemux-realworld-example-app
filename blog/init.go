package blog

import (
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
)

func init() {
	g := rwe.API.WithMiddleware(org.UserMiddleware)

	g.GET("/tags/", listTagsHandler)
	g.GET("/articles", listArticlesHandler)
	g.GET("/articles/feed", articleFeedHandler)
	g.GET("/articles/:slug", showArticleHandler)
	g.GET("/articles/:slug/comments", listCommentsHandler)
	g.GET("/articles/:slug/comments/:id", showCommentHandler)

	g = g.WithMiddleware(org.MustUserMiddleware)

	g.POST("/articles", createArticleHandler)
	g.PUT("/articles/:slug", updateArticleHandler)
	g.DELETE("/articles/:slug", deleteArticleHandler)

	g.POST("/articles/:slug/favorite", favoriteArticleHandler)
	g.DELETE("/articles/:slug/favorite", unfavoriteArticleHandler)

	g.POST("/articles/:slug/comments", createCommentHandler)
	g.DELETE("/articles/:slug/comments/:id", deleteCommentHandler)
}

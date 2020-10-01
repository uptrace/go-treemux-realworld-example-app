package blog

import (
	"errors"
	"net/http"

	"github.com/uptrace/go-realworld-example-app/httputil"
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
	"github.com/vmihailenco/treemux"
)

func listCommentsEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	article, err := SelectArticle(ctx, req.Param("slug"))
	if err != nil {
		return err
	}

	var userID uint64
	if user := org.UserFromContext(ctx); user != nil {
		userID = user.ID
	}

	comments := make([]*Comment, 0)
	if err := rwe.PGMain().ModelContext(ctx, &comments).
		ColumnExpr("c.*").
		Relation("Author").
		Apply(authorFollowingColumn(userID)).
		Where("article_id = ?", article.ID).
		Select(); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"comments": comments,
	})
}

func showCommentEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	article, err := SelectArticle(ctx, req.Param("slug"))
	if err != nil {
		return err
	}

	id, err := req.Params.Uint64("id")
	if err != nil {
		return err
	}

	var userID uint64
	if user := org.UserFromContext(ctx); user != nil {
		userID = user.ID
	}

	comment := new(Comment)
	if err := rwe.PGMain().ModelContext(ctx, comment).
		ColumnExpr("c.*").
		Relation("Author").
		Apply(authorFollowingColumn(userID)).
		Where("c.id = ?", id).
		Where("article_id = ?", article.ID).
		Select(); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"comment": comment,
	})
}

func createCommentEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	article, err := SelectArticle(ctx, req.Param("slug"))
	if err != nil {
		return err
	}

	var in struct {
		Comment *Comment `json:"comment"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.Comment == nil {
		return errors.New(`JSON field "comment" is required`)
	}

	comment := in.Comment

	comment.AuthorID = user.ID
	comment.ArticleID = article.ID
	comment.CreatedAt = rwe.Clock.Now()
	comment.UpdatedAt = rwe.Clock.Now()

	if _, err := rwe.PGMain().
		ModelContext(ctx, comment).
		Insert(); err != nil {
		return err
	}

	comment.Author = org.NewProfile(user)
	return httputil.Write(w, httputil.M{
		"comment": comment,
	})
}

func deleteCommentEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	article, err := SelectArticle(ctx, req.Param("slug"))
	if err != nil {
		return err
	}

	if _, err := rwe.PGMain().
		ModelContext(ctx, (*Comment)(nil)).
		Where("author_id = ?", user.ID).
		Where("article_id = ?", article.ID).
		Delete(); err != nil {
		return err
	}

	return nil
}

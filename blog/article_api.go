package blog

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/uptrace/go-realworld-example-app/httputil"
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
	"github.com/vmihailenco/treemux"

	"github.com/go-pg/pg/v10"
	"github.com/gosimple/slug"
)

const (
	kb = 10
	mb = 20
)

func makeSlug(title string) string {
	return slug.Make(title) + "-" + strconv.Itoa(rand.Int())
}

func listArticlesEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(req)
	if err != nil {
		return err
	}

	articles := make([]*Article, 0)
	if err := rwe.PGMain().ModelContext(ctx, &articles).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		Select(); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"articles":      articles,
		"articlesCount": len(articles),
	})
}

func showArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, f)
	if err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"article": article,
	})
}

func articleFeedEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(req)
	if err != nil {
		return err
	}
	f.Feed = true

	articles := make([]*Article, 0)
	if err := rwe.PGMain().
		ModelContext(ctx, &articles).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Select(); err != nil {
		return err
	}

	return httputil.Write(w, httputil.M{
		"articles":      articles,
		"articlesCount": len(articles),
	})
}

func createArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	var in struct {
		Article *Article `json:"article"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 100<<kb); err != nil {
		return err
	}

	if in.Article == nil {
		return errors.New(`JSON field "article" is required`)
	}

	article := in.Article

	article.Slug = makeSlug(article.Title)
	article.AuthorID = user.ID
	article.CreatedAt = rwe.Clock.Now()
	article.UpdatedAt = rwe.Clock.Now()

	if _, err := rwe.PGMain().
		ModelContext(ctx, article).
		Insert(); err != nil {
		return err
	}

	if err := createTags(ctx, article); err != nil {
		return err
	}

	article.Author = org.NewProfile(user)
	return httputil.Write(w, httputil.M{
		"article": article,
	})
}

func updateArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	var in struct {
		Article *Article `json:"article"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 100<<kb); err != nil {
		return err
	}

	if in.Article == nil {
		return errors.New(`JSON field "article" is required`)
	}

	article := in.Article

	if _, err := rwe.PGMain().
		ModelContext(ctx, article).
		Set("title = ?", article.Title).
		Set("description = ?", article.Description).
		Set("body = ?", article.Body).
		Set("updated_at = ?", rwe.Clock.Now()).
		Where("slug = ?", req.Param("slug")).
		Returning("*").
		Update(); err != nil {
		return err
	}

	if _, err := rwe.PGMain().ModelContext(ctx, (*ArticleTag)(nil)).
		Where("article_id = ?", article.ID).
		Delete(); err != nil {
		return err
	}

	if err := createTags(ctx, article); err != nil {
		return err
	}

	if article.TagList == nil {
		article.TagList = make([]string, 0)
	}

	article.Author = org.NewProfile(user)
	return httputil.Write(w, httputil.M{
		"article": article,
	})
}

func deleteArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	if _, err := rwe.PGMain().
		ModelContext(ctx, (*Article)(nil)).
		Where("author_id = ?", user.ID).
		Where("slug = ?", req.Param("slug")).
		Delete(); err != nil {
		return err
	}

	return nil
}

func createTags(ctx context.Context, article *Article) error {
	if len(article.TagList) == 0 {
		return nil
	}

	tags := make([]ArticleTag, 0, len(article.TagList))
	for _, t := range article.TagList {
		tags = append(tags, ArticleTag{
			ArticleID: article.ID,
			Tag:       t,
		})
	}

	if _, err := rwe.PGMain().
		ModelContext(ctx, &tags).
		Insert(); err != nil {
		return err
	}

	return nil
}

func favoriteArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f, err := decodeArticleFilter(req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, f)
	if err != nil {
		return err
	}

	favoriteArticle := &FavoriteArticle{
		UserID:    user.ID,
		ArticleID: article.ID,
	}
	res, err := rwe.PGMain().
		ModelContext(ctx, favoriteArticle).
		Insert()
	if err != nil {
		return err
	}

	if res.RowsAffected() != 0 {
		article.Favorited = true
		article.FavoritesCount = article.FavoritesCount + 1
	}

	return httputil.Write(w, httputil.M{
		"article": article,
	})
}

func unfavoriteArticleEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f, err := decodeArticleFilter(req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, f)
	if err != nil {
		return err
	}

	res, err := rwe.PGMain().
		ModelContext(ctx, (*FavoriteArticle)(nil)).
		Where("user_id = ?", user.ID).
		Where("article_id = ?", article.ID).
		Delete()
	if err != nil {
		return err
	}

	if res.RowsAffected() != 0 {
		article.Favorited = false
		article.FavoritesCount = article.FavoritesCount - 1
	}

	return httputil.Write(w, httputil.M{
		"article": article,
	})
}

func listTagsEndpoint(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	tags := make([]string, 0)
	if err := rwe.PGMain().ModelContext(ctx, (*ArticleTag)(nil)).
		ColumnExpr("tag").
		GroupExpr("tag").
		OrderExpr("count(tag) DESC").
		Select(&tags); err != nil && err != pg.ErrNoRows {
		return err
	}

	return httputil.Write(w, httputil.M{
		"tags": tags,
	})
}

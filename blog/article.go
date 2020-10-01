package blog

import (
	"context"
	"time"

	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
)

type Article struct {
	tableName struct{} `pg:"articles,alias:a"`

	ID          uint64 `json:"-"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`

	Author   *org.Profile `json:"author" pg:"rel:has-one"`
	AuthorID uint64       `json:"-"`

	Tags    []ArticleTag `json:"-" pg:"rel:has-many"`
	TagList []string     `json:"tagList" pg:"-,array"`

	Favorited      bool `json:"favorited" pg:"-"`
	FavoritesCount int  `json:"favoritesCount" pg:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ArticleTag struct {
	tableName struct{} `pg:"alias:t"`

	ArticleID uint64
	Tag       string
}

type FavoriteArticle struct {
	tableName struct{} `pg:"alias:fa"`

	UserID    uint64
	ArticleID uint64
}

func SelectArticle(c context.Context, slug string) (*Article, error) {
	article := new(Article)
	if err := rwe.PGMain().ModelContext(c, article).
		Where("slug = ?", slug).
		Select(); err != nil {
		return nil, err
	}
	return article, nil
}

func selectArticleByFilter(ctx context.Context, f *ArticleFilter) (*Article, error) {
	article := new(Article)
	if err := rwe.PGMain().
		ModelContext(ctx, article).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Select(); err != nil {
		return nil, err
	}

	if article.TagList == nil {
		article.TagList = make([]string, 0)
	}

	return article, nil
}

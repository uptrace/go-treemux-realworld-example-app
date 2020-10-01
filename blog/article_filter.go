package blog

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/go-pg/urlstruct"
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
	"github.com/vmihailenco/treemux"
)

type ArticleFilter struct {
	UserID    uint64
	Author    string
	Tag       string
	Favorited string
	Slug      string
	Feed      bool
	urlstruct.Pager
}

func decodeArticleFilter(req treemux.Request) (*ArticleFilter, error) {
	ctx := req.Context()
	query := req.URL.Query()

	f := &ArticleFilter{
		Tag:       query.Get("tag"),
		Author:    query.Get("author"),
		Favorited: query.Get("favorited"),
		Slug:      req.Param("slug"),
	}

	if user := org.UserFromContext(ctx); user != nil {
		f.UserID = user.ID
	}

	return f, nil
}

func (f *ArticleFilter) query(q *orm.Query) (*orm.Query, error) {
	q = q.Relation("Author")

	{
		subq := pg.Model((*ArticleTag)(nil)).
			ColumnExpr("array_agg(t.tag)::text[]").
			Where("t.article_id = a.id")

		q = q.ColumnExpr("(?) AS tag_list", subq)
	}

	if f.UserID == 0 {
		q = q.ColumnExpr("false AS favorited")
	} else {
		subq := pg.Model((*FavoriteArticle)(nil)).
			Where("fa.article_id = a.id").
			Where("fa.user_id = ?", f.UserID)

		q = q.ColumnExpr("EXISTS (?) AS favorited", subq)
	}

	q.Apply(authorFollowingColumn(f.UserID))

	{
		subq := pg.Model((*FavoriteArticle)(nil)).
			ColumnExpr("count(*)").
			Where("fa.article_id = a.id")

		q = q.ColumnExpr("(?) AS favorites_count", subq)
	}

	if f.Author != "" {
		q = q.Where("author.username = ?", f.Author)
	}

	if f.Tag != "" {
		subq := pg.Model((*ArticleTag)(nil)).
			Distinct().
			ColumnExpr("t.article_id").
			Where("t.tag = ?", f.Tag)

		q = q.Where("a.id IN (?)", subq)
	}

	if f.Feed {
		subq := pg.Model((*org.FollowUser)(nil)).
			ColumnExpr("fu.followed_user_id").
			Where("fu.user_id = ?", f.UserID)

		q = q.Where("a.author_id IN (?)", subq)
	} else if f.Slug != "" {
		q = q.Where("a.slug = ?", f.Slug)
	}

	return q, nil
}

func authorFollowingColumn(userID uint64) func(*orm.Query) (*orm.Query, error) {
	return func(q *orm.Query) (*orm.Query, error) {
		if userID == 0 {
			q = q.ColumnExpr("false AS author__following")
		} else {
			subq := rwe.PGMain().Model((*org.FollowUser)(nil)).
				Where("fu.followed_user_id = author_id").
				Where("fu.user_id = ?", userID)

			q = q.ColumnExpr("EXISTS (?) AS author__following", subq)
		}

		return q, nil
	}
}

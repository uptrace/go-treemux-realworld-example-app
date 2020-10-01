package blog

import (
	"time"

	"github.com/uptrace/go-realworld-example-app/org"
)

type Comment struct {
	tableName struct{} `pg:"comments,alias:c"`

	ID   uint64 `json:"id"`
	Body string `json:"body"`

	Author   *org.Profile `json:"author" pg:"rel:has-one"`
	AuthorID uint64       `json:"-"`

	ArticleID uint64 `json:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

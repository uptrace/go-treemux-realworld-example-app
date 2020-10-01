package blog_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
	. "github.com/uptrace/go-realworld-example-app/testbed"
	"github.com/uptrace/go-realworld-example-app/xconfig"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "blog")
}

var ctx context.Context

func init() {
	mock := clock.NewMock()
	mock.Set(time.Date(2020, time.January, 1, 2, 3, 4, 5000, time.UTC))
	rwe.Clock = mock

	ctx = context.Background()

	cfg, err := xconfig.LoadConfig("test")
	if err != nil {
		panic(err)
	}

	ctx = rwe.Init(ctx, cfg)
}

var _ = Describe("createArticle", func() {
	var data map[string]interface{}
	var slug string
	var user *org.User

	var helloArticleKeys, fooArticleKeys, favoritedArticleKeys Keys

	createFollowedUser := func() *org.User {
		followedUser := &org.User{
			Username:     "FollowedUser",
			Email:        "foo@bar.com",
			PasswordHash: "h2",
		}
		_, err := rwe.PGMain().Model(followedUser).Insert()
		Expect(err).NotTo(HaveOccurred())

		url := fmt.Sprintf("/api/profiles/%s/follow", followedUser.Username)
		resp := PostWithToken(url, "", user.ID)
		_ = ParseJSON(resp, 200)

		return followedUser
	}

	BeforeEach(func() {
		ResetAll(ctx)

		helloArticleKeys = Keys{
			"title":          Equal("Hello world"),
			"slug":           HavePrefix("hello-world-"),
			"description":    Equal("Hello world article description!"),
			"body":           Equal("Hello world article body."),
			"author":         Equal(map[string]interface{}{"following": false, "username": "CurrentUser", "bio": "", "image": ""}),
			"tagList":        ConsistOf([]interface{}{"greeting", "welcome", "salut"}),
			"favoritesCount": Equal(float64(0)),
			"favorited":      Equal(false),
			"createdAt":      Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
			"updatedAt":      Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
		}

		favoritedArticleKeys = ExtendKeys(helloArticleKeys, Keys{
			"favorited":      Equal(true),
			"favoritesCount": Equal(float64(1)),
		})

		fooArticleKeys = Keys{
			"title":          Equal("Foo bar"),
			"slug":           HavePrefix("foo-bar-"),
			"description":    Equal("Foo bar article description!"),
			"body":           Equal("Foo bar article body."),
			"author":         Equal(map[string]interface{}{"following": false, "username": "CurrentUser", "bio": "", "image": ""}),
			"tagList":        ConsistOf([]interface{}{"foobar", "variable"}),
			"favoritesCount": Equal(float64(0)),
			"favorited":      Equal(false),
			"createdAt":      Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
			"updatedAt":      Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
		}

		user = &org.User{
			Username:     "CurrentUser",
			Email:        "hello@world.com",
			PasswordHash: "#1",
		}
		_, err := rwe.PGMain().Model(user).Insert()
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		json := `{"article": {"title": "Hello world", "description": "Hello world article description!", "body": "Hello world article body.", "tagList": ["greeting", "welcome", "salut"]}}`
		resp := PostWithToken("/api/articles", json, user.ID)

		data = ParseJSON(resp, http.StatusOK)
		slug = data["article"].(map[string]interface{})["slug"].(string)
	})

	It("creates new article", func() {
		Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
	})

	Describe("showFeed", func() {
		BeforeEach(func() {
			followedUser := createFollowedUser()

			json := `{"article": {"title": "Foo bar", "description": "Foo bar article description!", "body": "Foo bar article body.", "tagList": ["foobar", "variable"]}}`
			resp := PostWithToken("/api/articles", json, followedUser.ID)

			_ = ParseJSON(resp, http.StatusOK)

			resp = GetWithToken("/api/articles/feed", user.ID)
			data = ParseJSON(resp, http.StatusOK)
		})

		It("returns article", func() {
			articles := data["articles"].([]interface{})

			Expect(articles).To(HaveLen(1))
			followedAuthorKeys := ExtendKeys(fooArticleKeys, Keys{
				"author": Equal(map[string]interface{}{
					"following": true,
					"username":  "FollowedUser",
					"bio":       "",
					"image":     "",
				}),
			})
			Expect(articles[0].(map[string]interface{})).To(MatchAllKeys(followedAuthorKeys))
		})
	})

	Describe("showArticle", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s", slug)
			resp := Get(url)

			data = ParseJSON(resp, http.StatusOK)
		})

		It("returns article", func() {
			Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
		})
	})

	Describe("listArticles", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s?author=CurrentUser", slug)
			resp := Get(url)

			data = ParseJSON(resp, http.StatusOK)
		})

		It("returns articles by author", func() {
			Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
		})
	})

	Describe("favoriteArticle", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s/favorite", slug)
			resp := PostWithToken(url, "", user.ID)
			_ = ParseJSON(resp, 200)

			url = fmt.Sprintf("/api/articles/%s", slug)
			resp = GetWithToken(url, user.ID)
			data = ParseJSON(resp, 200)
		})

		It("returns favorited article", func() {
			Expect(data["article"]).To(MatchAllKeys(favoritedArticleKeys))
		})

		Describe("unfavoriteArticle", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/favorite", slug)
				resp := DeleteWithToken(url, user.ID)
				_ = ParseJSON(resp, 200)

				url = fmt.Sprintf("/api/articles/%s", slug)
				resp = GetWithToken(url, user.ID)
				data = ParseJSON(resp, 200)
			})

			It("returns article", func() {
				Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
			})
		})
	})

	Describe("listArticles", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s/favorite", slug)
			resp := PostWithToken(url, "", user.ID)
			_ = ParseJSON(resp, 200)

			resp = GetWithToken("/api/articles", user.ID)
			data = ParseJSON(resp, 200)
		})

		It("returns articles", func() {
			articles := data["articles"].([]interface{})

			Expect(articles).To(HaveLen(1))
			article := articles[0].(map[string]interface{})
			Expect(article).To(MatchAllKeys(favoritedArticleKeys))
		})
	})

	Describe("updateArticle", func() {
		BeforeEach(func() {
			json := `{"article": {"title": "Foo bar", "description": "Foo bar article description!", "body": "Foo bar article body.", "tagList": []}}`

			url := fmt.Sprintf("/api/articles/%s", slug)
			resp := PutWithToken(url, json, user.ID)
			data = ParseJSON(resp, 200)
		})

		It("returns article", func() {
			updatedArticleKeys := ExtendKeys(fooArticleKeys, Keys{
				"slug":      HavePrefix("hello-world-"),
				"tagList":   Equal([]interface{}{}),
				"updatedAt": Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
			})
			Expect(data["article"]).To(MatchAllKeys(updatedArticleKeys))
		})
	})

	Describe("deleteArticle", func() {
		var resp *httptest.ResponseRecorder

		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s", slug)
			resp = DeleteWithToken(url, user.ID)
		})

		It("deletes article", func() {
			Expect(resp.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("createComment", func() {
		var commentKeys Keys
		var commentID uint64
		var followedUser *org.User

		BeforeEach(func() {
			commentKeys = Keys{
				"id":        Not(BeZero()),
				"body":      Equal("First comment."),
				"author":    Equal(map[string]interface{}{"following": false, "username": "FollowedUser", "bio": "", "image": ""}),
				"createdAt": Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
				"updatedAt": Equal(rwe.Clock.Now().Format(time.RFC3339Nano)),
			}

			followedUser = createFollowedUser()

			json := `{"comment": {"body": "First comment."}}`
			url := fmt.Sprintf("/api/articles/%s/comments", slug)
			resp := PostWithToken(url, json, followedUser.ID)
			data = ParseJSON(resp, 200)

			commentID = uint64(data["comment"].(map[string]interface{})["id"].(float64))
		})

		It("returns created comment to article", func() {
			Expect(data["comment"]).To(MatchAllKeys(commentKeys))
		})

		Describe("showComment", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp := Get(url)
				data = ParseJSON(resp, 200)
			})

			It("returns comment to article", func() {
				Expect(data["comment"]).To(MatchAllKeys(commentKeys))
			})
		})

		Describe("showComment with authentication", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp := GetWithToken(url, user.ID)
				data = ParseJSON(resp, 200)
			})

			It("returns comment to article", func() {
				followedCommentKeys := ExtendKeys(commentKeys, Keys{
					"author": Equal(map[string]interface{}{"following": true, "username": "FollowedUser", "bio": "", "image": ""}),
				})
				Expect(data["comment"]).To(MatchAllKeys(followedCommentKeys))
			})
		})

		Describe("listArticleComments", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments", slug)
				resp := GetWithToken(url, user.ID)
				data = ParseJSON(resp, 200)
			})

			It("returns article comments", func() {
				followedCommentKeys := ExtendKeys(commentKeys, Keys{
					"author": Equal(map[string]interface{}{"following": true, "username": "FollowedUser", "bio": "", "image": ""}),
				})
				Expect(data["comments"].([]interface{})[0]).To(MatchAllKeys(followedCommentKeys))
			})
		})

		Describe("deleteComment", func() {
			var resp *httptest.ResponseRecorder

			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp = DeleteWithToken(url, followedUser.ID)
			})

			It("deletes comment", func() {
				Expect(resp.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("listTags", func() {
		BeforeEach(func() {
			resp := Get("/api/tags/")
			data = ParseJSON(resp, 200)
		})

		It("returns tags", func() {
			Expect(data["tags"]).To(ConsistOf([]string{
				"greeting",
				"salut",
				"welcome",
			}))
		})
	})
})

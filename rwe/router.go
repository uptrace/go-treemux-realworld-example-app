package rwe

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-redis/redis_rate/v9"
	"github.com/uptrace/go-realworld-example-app/httputil/httperror"
	"github.com/vmihailenco/treemux"
	"github.com/vmihailenco/treemux/extra/reqlog"
	"github.com/vmihailenco/treemux/extra/treemuxgzip"
	"github.com/vmihailenco/treemux/extra/treemuxotel"
)

var (
	Router *treemux.TreeMux
	API    *treemux.Group
)

func init() {
	Router = treemux.New(
		treemux.WithMiddleware(treemuxgzip.Middleware),
		treemux.WithMiddleware(treemuxotel.Middleware),
		treemux.WithMiddleware(reqlog.Middleware),
		treemux.WithErrorHandler(errorHandler),
	)

	API = Router.NewGroup("/api",
		treemux.WithMiddleware(corsMiddleware),
		treemux.WithMiddleware(rateLimitMiddleware),
	)
}

func errorHandler(w http.ResponseWriter, req treemux.Request, err error) {
	httpErr := httperror.From(err)
	_ = treemux.JSON(w, httpErr)
}

func corsMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			return next(w, req)
		}

		h := w.Header()

		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")

		// CORS preflight.
		if req.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,HEAD")
			h.Set("Access-Control-Allow-Headers", "authorization,content-type")
			h.Set("Access-Control-Max-Age", "86400")
			return nil
		}

		return next(w, req)
	}
}

func rateLimitMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	limit := redis_rate.PerMinute(100)

	return func(w http.ResponseWriter, req treemux.Request) error {
		if req.Method == http.MethodOptions {
			return next(w, req)
		}

		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return err
		}

		rateKey := "rl:" + host
		res, err := RateLimiter().Allow(req.Context(), rateKey, limit)
		if err != nil {
			return err
		}
		if res.Allowed == 0 {
			return errors.New("rate limited")
		}

		return next(w, req)
	}
}

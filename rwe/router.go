package rwe

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-redis/redis_rate/v9"
	"github.com/uptrace/go-realworld-example-app/httputil/httperror"
	"github.com/vmihailenco/treemux"
	"github.com/vmihailenco/treemux/extra/reqlog"
)

var (
	Router = treemux.New()
	API    *treemux.LockedGroup
)

func init() {
	Router.Use(reqlog.Middleware)

	api := Router.NewGroup("/api")
	API = api.Lock()

	Router.ErrorHandler = func(w http.ResponseWriter, req treemux.Request, err error) {
		httpErr := httperror.From(err)
		_ = treemux.JSON(w, httpErr)
	}

	api.Use(corsMiddleware)
	api.Use(rateLimitMiddleware)
	// Router.Use(gintrace.Middleware("rwe"))

	api.OPTIONS("/*", corsPreflight)

	API = api.Lock()
}

func corsPreflight(w http.ResponseWriter, req treemux.Request) error {
	h := w.Header()
	if origin := req.Header.Get("Origin"); origin != "" {
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,HEAD")
		h.Set("Access-Control-Allow-Headers", "authorization,content-type")
		h.Set("Access-Control-Max-Age", "86400")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func corsMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		if origin := req.Header.Get("Origin"); origin != "" {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
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

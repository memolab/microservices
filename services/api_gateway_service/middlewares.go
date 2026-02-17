package main

import (
	"context"
	"fmt"
	"log/slog"
	"microservices-demo/common/rid"
	"microservices-demo/common/types"
	"net/http"
	"time"
)

type Middleware func(h http.Handler) http.Handler

func applyMiddlewares(nfs ...Middleware) Middleware {
	return func(nf http.Handler) http.Handler {
		for i := len(nfs) - 1; i >= 0; i-- {
			nf = nfs[i](nf)
		}
		return nf
	}
}

func loggingMiddleware(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rid := rid.New8()
		ctx := context.WithValue(r.Context(), types.RequestIDKey{}, rid)

		defer func() {
			var cid string
			if id, ok := r.Context().Value(types.ConnIDKey{}).(string); ok {
				cid = id
			}
			if err := recover(); err != nil {
				slog.Error("request recover", "error", err.(error),
					"user-agent", r.Header.Get("user-agent"),
					"client-ip", r.Header.Get("x-forwarded-for"),
					"path", r.URL.Path,
					"d", time.Since(t),
					"cid", cid,
					"rid", rid)
			} else {
				slog.Debug("request",
					"user-agent", r.Header.Get("user-agent"),
					"client-ip", r.Header.Get("x-forwarded-for"),
					"path", r.URL.Path,
					"d", time.Since(t),
					"cid", cid,
					"rid", rid)
			}
		}()
		f.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setHeadersMiddleware(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid, _ := r.Context().Value(types.RequestIDKey{}).(string)
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("content-type", "text/html; charset=utf-8")
		w.Header().Set("cache-control", "no-cache, no-store, private")
		w.Header().Set("expires", "-1")
		w.Header().Set("vary", "Accept-Encoding,Origin")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("request-id", rid)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		f.ServeHTTP(w, r)
	})
}

func setAssetsHeadersMiddleware(f http.Handler, maxAge time.Duration) http.Handler {
	val := fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", val)
		f.ServeHTTP(w, r)
	})
}

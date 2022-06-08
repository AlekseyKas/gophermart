package middlewarecustom

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/sirupsen/logrus"
)

type gzipBodyWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var cookie *http.Cookie
		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				cookie = cook
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}

		if cookie != nil {
			b, err := storage.DB.CheckCookie(c)
			switch {
			case err == nil:
				logrus.Info("Get cookie without err")
			case strings.Contains(err.Error(), "unexpected EOF") || strings.Contains(err.Error(), "failed"):
				rw.WriteHeader(http.StatusInternalServerError)
			}

			switch {
			case b:
				//already login to /login and /register
				if req.URL.Path == "/api/user/register" || req.URL.Path == "/api/user/login" {
					rw.WriteHeader(http.StatusOK)
					return
				} else {
					next.ServeHTTP(rw, req)
				}

			case !b:
				if req.URL.Path == "/api/user/register" || req.URL.Path == "/api/user/login" {
					next.ServeHTTP(rw, req)
				} else {
					rw.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		} else {
			if req.URL.Path == "/api/user/register" || req.URL.Path == "/api/user/login" {
				next.ServeHTTP(rw, req)
			} else {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	})
}

func CompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		// w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		next.ServeHTTP(gzipBodyWriter{
			ResponseWriter: w,
			writer:         gz,
		}, r)
	})
}

func DecompressGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gz.Close()
		r.Body = gz
		next.ServeHTTP(w, r)

	})
}

package middlewarecustom

import (
	"net/http"
	"strings"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/sirupsen/logrus"
)

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
					logrus.Info("11111111111111122222222222: ", b)
					next.ServeHTTP(rw, req)
				}

			case !b:
				if req.URL.Path == "/api/user/register" || req.URL.Path == "/api/user/login" {
					next.ServeHTTP(rw, req)
				} else {
					logrus.Info("111111111111111aaaaaaaaaaa: ", b)
					rw.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		} else {
			if req.URL.Path == "/api/user/register" || req.URL.Path == "/api/user/login" {
				next.ServeHTTP(rw, req)
			} else {
				logrus.Info("111111111111111bbbbbbbbbb")
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	})
}

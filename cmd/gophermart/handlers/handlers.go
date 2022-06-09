package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/neonxp/checksum"
	"github.com/neonxp/checksum/luhn"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/middlewarecustom"
)

type B struct {
	wg  *sync.WaitGroup
	ctx context.Context
	r   chi.Router
}

func NewArgs(r chi.Router, wg *sync.WaitGroup, ctx context.Context) *B {
	return &B{r: r, wg: wg, ctx: ctx}
}
func Router(r chi.Router) {

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewarecustom.CheckCookie)
	r.Use(middlewarecustom.CompressGzip)
	r.Use(middlewarecustom.DecompressGzip)

	//регистрация пользователя
	r.Post("/api/user/register", register())
	// аутентификация пользователя
	r.Post("/api/user/login", login())
	// загрузка пользователем номера заказа для расчёта
	r.Post("/api/user/orders", loadOrder())
	// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	r.Post("/api/user/balance/withdraw", withdrawOrder())
	// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	r.Get("/api/user/orders", getOrders())
	// получение текущего баланса счёта баллов лояльности пользователя
	r.Get("/api/user/balance", getBalance())
	// получение информации о выводе средств с накопительного счёта пользователем
	r.Get("/api/user/balance/withdrawals", getWithdraws())

}

func register() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logrus.Error("Error read body: ", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		u := storage.Users
		err = json.Unmarshal(out, &u)
		if err != nil {
			logrus.Error("Error unmarshal body: ", err)
			rw.WriteHeader(http.StatusBadRequest)
		}
		if u.Login == "" || u.Password == "" {
			logrus.Error("Wrong format of user or password.")
			rw.WriteHeader(http.StatusBadRequest)
		}
		err409 := errors.New("ERROR: duplicate key value violates unique constraint \"users_login_key\" (SQLSTATE 23505)")
		cookie, err := storage.DB.CreateUser(u, strings.Split(req.RemoteAddr, ":")[0])
		logrus.Info(cookie, err)
		switch {
		case err == nil:
			logrus.Info("User added: ", u.Login)
			http.SetCookie(rw, &http.Cookie{Name: cookie.Name, Value: cookie.Value, MaxAge: cookie.MaxAge, Expires: cookie.Expires})
			rw.WriteHeader(http.StatusOK)
		case err.Error() == err409.Error():
			logrus.Error("User already exist: ", u.Login)
			rw.WriteHeader(http.StatusConflict)
		default:
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}

//login users
func login() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logrus.Error("Error read body: ", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		u := storage.Users
		err = json.Unmarshal(out, &u)
		if err != nil {
			logrus.Error("Error unmarshal body: ", err)
			rw.WriteHeader(http.StatusBadRequest)
		}

		if u.Login == "" || u.Password == "" {
			logrus.Error("Wrong format of user or password.")
			rw.WriteHeader(http.StatusBadRequest)
		}
		cookie, err := storage.DB.AuthUser(u, strings.Split(req.RemoteAddr, ":")[0])
		err1 := errors.New("invalid password")
		err2 := errors.New("no rows in result set")
		switch {
		case err == nil:
			logrus.Info("User login: ", u.Login)
			http.SetCookie(rw, &http.Cookie{Name: cookie.Name, Value: cookie.Value, MaxAge: cookie.MaxAge, Expires: cookie.Expires})
			rw.WriteHeader(http.StatusOK)
		case err.Error() == err1.Error() || strings.Contains(err.Error(), err2.Error()):
			logrus.Error("Incorrect user or password.")
			rw.WriteHeader(http.StatusUnauthorized)
		default:
			rw.WriteHeader(http.StatusInternalServerError)
		}

	}
}

func loadOrder() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		if !strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		rw.Header().Add("Content-Type", "text/plain")
		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}
		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logrus.Error("Error read body for load order: ", err)
		}

		err = luhn.Check(string(out))
		switch err {
		case checksum.ErrInvalidNumber:
			logrus.Info("Error format order: ", string(out))
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case checksum.ErrInvalidChecksum:
			logrus.Info("Error format order: ", string(out))
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case nil:
			userID, err := storage.DB.CheckUser(c)
			if err != nil {
				logrus.Error("Faild check user: ", err)
			}
			err = storage.DB.LoadOrder(string(out), c, userID)
			if err != nil {
				logrus.Error("Error load order to DB: ", err)
			}
			switch {
			case err == nil:
				logrus.Info("Order regitred")
				rw.WriteHeader(http.StatusAccepted)
				logrus.Error(err)
			case strings.Contains(err.Error(), "orders_number_key"):
				logrus.Error("Order already exist and add other user")
				rw.WriteHeader(http.StatusConflict)
			case strings.Contains(err.Error(), "orders_user_id_number_key"):
				rw.WriteHeader(http.StatusOK)
			default:
				rw.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

func withdrawOrder() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}
		defer req.Body.Close()
		out, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logrus.Error("Faild read body withdraw: ", err)
		}
		var w storage.Withdraw
		err = json.Unmarshal(out, &w)
		if err != nil {
			logrus.Error("Error unmarshal withdraw: ", err)
		}
		logrus.Info(w.Order)
		err = luhn.Check(w.Order)
		switch err {
		case checksum.ErrInvalidNumber:
			logrus.Info("Withdraw format order wrong: ", w.Order)
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case checksum.ErrInvalidChecksum:
			logrus.Info("Withdraw format order wrong: ", w.Order)
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case nil:
			userID, err := storage.DB.CheckUser(c)
			if err != nil {
				logrus.Error("Faild check user: ", err)
			}
			b, err := storage.DB.UpdateWithdraw(w, userID)
			if err != nil {
				logrus.Error("Error update withdraw: ", err)
			}
			switch {
			case b:
				rw.Header().Add("Content-Type", "application.json")
				rw.WriteHeader(http.StatusOK)
			case !b:
				rw.Header().Add("Content-Type", "application.json")
				rw.WriteHeader(http.StatusPaymentRequired)
			}
		}
	}
}

func getWithdraws() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}
		userID, err := storage.DB.CheckUser(c)
		if err != nil {
			logrus.Error("Faild check user: ", err)
		}

		withdraws, err := storage.DB.Getwithdraws(userID)
		if err != nil {
			logrus.Error("Error get withdraws: ", err)
		}
		rw.Header().Add("Content-Type", "application/json")
		if len(withdraws) == 0 {
			rw.WriteHeader(http.StatusNoContent)
		} else {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			err := encoder.Encode(withdraws)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			rw.WriteHeader(http.StatusOK)
			rw.Write(buf.Bytes())
		}
	}
}
func getOrders() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}
		userID, err := storage.DB.CheckUser(c)
		if err != nil {
			logrus.Error("Faild check user: ", err)
		}

		orders, err := storage.DB.GetOrders(userID)
		if err != nil {
			logrus.Error(err)
		}
		if len(orders) == 0 {
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusNoContent)
		} else {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			err := encoder.Encode(orders)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write(buf.Bytes())
		}
	}
}

func getBalance() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var c storage.Cookie
		for _, cook := range req.Cookies() {
			if cook.Name == "gophermart" {
				c = storage.Cookie{
					Name:  cook.Name,
					Value: cook.Value,
				}
			}
		}
		userID, err := storage.DB.CheckUser(c)
		if err != nil {
			logrus.Error("Error check userID: ", err)
		}
		balance, err := storage.DB.GetBalance(userID)
		if err != nil {
			logrus.Error("Error check userID: ", err)
		}
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		err = encoder.Encode(balance)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(buf.Bytes())
	}
}

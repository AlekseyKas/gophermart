package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/config"
	"github.com/AlekseyKas/gophermart/internal/middlewarecustom"
	"github.com/go-resty/resty/v2"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/neonxp/checksum"
	"github.com/neonxp/checksum/luhn"
	"github.com/sirupsen/logrus"
)

type B struct {
	wg  *sync.WaitGroup
	ctx context.Context
	r   chi.Router
}

func NewArgs(r chi.Router, wg *sync.WaitGroup, ctx context.Context) *B {
	return &B{r: r, wg: wg, ctx: ctx}
}
func (args *B) Router(r chi.Router) {

	args.r.Use(middleware.RequestID)
	args.r.Use(middleware.RealIP)
	args.r.Use(middleware.Logger)
	args.r.Use(middleware.Recoverer)
	args.r.Use(middlewarecustom.CheckCookie)

	//регистрация пользователя
	args.r.Post("/api/user/register", register())
	// аутентификация пользователя
	args.r.Post("/api/user/login", login())
	// загрузка пользователем номера заказа для расчёта
	args.r.Post("/api/user/orders", args.loadOrder())
	// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	args.r.Post("/api/user/balance/withdraw", withdrawOrder())
	// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	args.r.Get("/api/user/orders", getOrders())
	// получение текущего баланса счёта баллов лояльности пользователя
	args.r.Get("/api/user/balance", getBalance())
	// получение информации о выводе средств с накопительного счёта пользователем
	args.r.Get("/api/user/balance/withdrawals", withdraw())

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
		cookie, err := storage.DB.CreateUser(u)
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
		cookie, err := storage.DB.AuthUser(u)
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

func (args *B) loadOrder() http.HandlerFunc {
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
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case checksum.ErrInvalidChecksum:
			rw.WriteHeader(http.StatusUnprocessableEntity)
		case nil:
			err := storage.DB.LoadOrder(string(out), c)
			if err != nil {
				logrus.Error("Error load order to DB: ", err)
			}
			switch {
			case err == nil:
				logrus.Info("Order regitred")
				rw.WriteHeader(http.StatusAccepted)
				args.wg.Add(1)
				go sendAccural(args.wg, args.ctx, out)
				logrus.Error(err)
			case strings.Contains(err.Error(), "orders_number_key"):
				logrus.Error("Order already exist and add other user")
				rw.WriteHeader(http.StatusConflict)
			case strings.Contains(err.Error(), "orders_user_id_number_key"):
				rw.WriteHeader(http.StatusOK)
			default:
				rw.WriteHeader(http.StatusInternalServerError)
			}
			// rw.WriteHeader(http.StatusOK)
		}
	}
}

func sendAccural(wg *sync.WaitGroup, ctx context.Context, number []byte) {
	//for test
	// client := resty.New()
	// _, err := client.R().
	// 	SetHeader("Content-Type", "application/json").
	// 	SetBody(number).
	// 	Post("http://" + config.Arg.SystemAddress + "/api/orders")
	// if err != nil {
	// 	logrus.Error(err)
	// }
	//for test
	client := resty.New()
	var status string
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Check status down")
			wg.Done()
			return
		case <-time.After(time.Second * 2):
			defer wg.Done()
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				Get("http://" + config.Arg.SystemAddress + "/api/orders/" + string(number))
			if err != nil {
				logrus.Error(err)
			}
			order := storage.Orders
			err = json.Unmarshal(resp.Body(), &order)
			if err != nil {
				logrus.Error("Error unmarshal order from accrual: ", err)
			}
			if order.Status != status {
				err = storage.DB.UpdateOrder(order)
				if err != nil {
					logrus.Error("Error update order in DB: ", err)
				}
				status = order.Status
				if status == "INVALID" || status == "PROCESSED" {
					return
				}
			}
		}
	}
}

func withdrawOrder() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

	}
}

func withdraw() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

	}
}
func getOrders() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

	}
}

func getBalance() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		logrus.Info("balance")
	}
}

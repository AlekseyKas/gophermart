package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/database"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage/migrations"
	"github.com/AlekseyKas/gophermart/internal/app"
	"github.com/AlekseyKas/gophermart/internal/config/migrate"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type Withdraw struct {
	Order       string    `json:"order,omitempty"`
	Sum         float64   `json:"sum,omitempty"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}

type Order struct {
	OrderID    int
	UserID     int       `json:"userID,omitempty"`
	Order      string    `json:"order,omitempty"`
	Status     string    `json:"status,omitempty"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"UploadedAt,omitempty"`
}
type OrderOutput struct {
	Order      string    `json:"number,omitempty"`
	Status     string    `json:"status,omitempty"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at,omitempty"`
}

type Cookie struct {
	Name    string    `json:"Name,omitempty"`
	Value   string    `json:"Value,omitempty"`
	Path    string    `json:"Path,omitempty"`
	MaxAge  int       `json:"MaxAge,omitempty"`
	Expires time.Time `json:"Expires,omitempty"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Cookie   Cookie `json:"Cookie,omitempty"`
}

type Database struct {
	Con   *pgxpool.Pool
	Loger logrus.FieldLogger
	DBURL string
	Ctx   context.Context
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// type Order struct {
// 	userID      int
// 	number      string
// 	status      string
// 	accrual     float64
// 	UploadedAt time.Time
// }

var DB Database
var IDB Storage
var Users User
var Orders Order
var Withdraws Withdraw

type Storage interface {
	InitDB(ctx context.Context, DBURL string) error
	CreateUser(u User) (cookie Cookie, err error)
	CheckCookie(c Cookie) (bool, error)
	// GetUser(u User) (string, error)
	AuthUser(u User) (Cookie, error)
	// ReconnectDB() error
	LoadOrder(number string, c Cookie, userID int) error
	UpdateOrder(order Order) error
	GetOrders(userID int) (Ords []OrderOutput, err error)
	UpdateWithdraw(w Withdraw, userID int) (b bool, err error)
	CheckUser(c Cookie) (userID int, err error)
	Getwithdraws(userID int) (withdr []Withdraw, err error)
	GetBalance(userID int) (balance Balance, err error)
}

func (d *Database) GetBalance(userID int) (balance Balance, err error) {
	var withdraw float64
	var withdraws float64

	row := d.Con.QueryRow(d.Ctx, "SELECT balance FROM balance")
	err = row.Scan(&balance.Current)
	if err != nil {
		d.Loger.Error("Error scan row get user by cookie: ", err)
		return balance, err
	}
	rows, err := d.Con.Query(d.Ctx, "SELECT withdraws FROM withdraws WHERE user_id = $1", userID)
	if err != nil {
		logrus.Error("Error select orders: ", err)
	}
	for rows.Next() {
		err = rows.Scan(&withdraw)
		if err != nil {
			logrus.Error("Error scan orders: ", err)
		}
		withdraws += withdraw
	}
	balance.Withdrawn = withdraws
	return balance, err
}

func (d *Database) Getwithdraws(userID int) (withdr []Withdraw, err error) {
	var w Withdraw
	rows, err := d.Con.Query(d.Ctx, "SELECT ordername, withdraws, processed_at FROM withdraws WHERE user_id = $1 order by processed_at", userID)
	if err != nil {
		logrus.Error("Error select orders: ", err)
	}
	for rows.Next() {
		err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			logrus.Error("Error scan orders: ", err)
		}
		withdr = append(withdr, w)
	}
	return withdr, err
}

func (d *Database) CheckUser(c Cookie) (userID int, err error) {
	var login string
	var password string
	var cookie Cookie

	row := d.Con.QueryRow(d.Ctx, "SELECT * FROM users WHERE cookie->>'Value' = $1", c.Value)
	err = row.Scan(&userID, &login, &password, &cookie)
	if err != nil {
		d.Loger.Error("Error scan row get user by cookie: ", err)
		return userID, err
	}
	return userID, nil
}

func (d *Database) UpdateWithdraw(w Withdraw, userID int) (b bool, err error) {
	var balance float64
	row, err := d.Con.Query(d.Ctx, "SELECT balance FROM balance WHERE user_id = $1", userID)
	if err != nil {
		logrus.Error("Error select balance: ", err)
	}
	for row.Next() {
		err = row.Scan(&balance)
		if err != nil {
			logrus.Error("Error scan orders: ", err)
		}
	}
	if balance < w.Sum {
		return b, err
	} else {
		bal := balance - w.Sum
		logrus.Info("WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW", w.Sum)
		_, err = d.Con.Exec(d.Ctx, "UPDATE balance SET balance = $1 WHERE user_id = $2;", bal, userID)
		if err != nil {
			logrus.Error("Error update balance: ", err)
		}
		_, err = d.Con.Exec(d.Ctx, "INSERT INTO withdraws (user_id, ordername, withdraws, processed_at) VALUES($1,$2, $3, $4)", userID, w.Order, w.Sum, time.Now())
		b = true
	}
	return b, err
}

func (d *Database) GetOrders(userID int) (Ords []OrderOutput, err error) {
	var order OrderOutput
	rows, err := d.Con.Query(d.Ctx, "SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 order by uploaded_at", userID)
	if err != nil {
		logrus.Error("Error select orders: ", err)
	}
	for rows.Next() {
		err = rows.Scan(&order.Order, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			logrus.Error("Error scan orders: ", err)
		}
		Ords = append(Ords, order)
	}
	return Ords, err
}

func (d *Database) UpdateOrder(order Order) error {
	_, err := d.Con.Exec(d.Ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3;", order.Status, order.Accrual, order.Order)
	if err != nil {
		logrus.Error("Error update accrual: ", err)
	}
	return err
}

func (d *Database) LoadOrder(number string, c Cookie, userID int) error {
	_, err := d.Con.Exec(d.Ctx, "INSERT INTO orders (user_id, number, uploaded_at) VALUES($1,$2, $3)", userID, number, time.Now())
	return err
}

func (d *Database) CheckCookie(c Cookie) (bool, error) {

	var login string
	var password string
	var cookie Cookie
	var userID int

	row := d.Con.QueryRow(d.Ctx, "SELECT * FROM users WHERE cookie->>'Value' = $1", c.Value)
	err := row.Scan(&userID, &login, &password, &cookie)
	if err != nil {
		d.Loger.Error("Error scan row: ", err)
		return false, err
	}
	if cookie.Expires.After(time.Now()) {
		return true, err
	}
	return false, err
}

func (d *Database) AuthUser(u User) (Cookie, error) {
	var cookie Cookie
	var login string
	var password string
	var userID int
	row := d.Con.QueryRow(d.Ctx, "SELECT * FROM users WHERE login = $1", u.Login)
	err := row.Scan(&userID, &login, &password, &cookie)
	if err != nil {
		d.Loger.Error("Error scan row: ", err)
	}
	res := app.CheckPasswordHash(u.Password, password)
	if res {
		switch d.Con {
		case nil:
			return cookie, err
		default:
			valhashDB := hmac.New(sha256.New, []byte(login+password))
			cookieDB := fmt.Sprintf("%x", valhashDB.Sum(nil))
			cookie := Cookie{
				Name:    "gophermart",
				Value:   cookieDB,
				MaxAge:  86400,
				Expires: time.Now().Local().Add(time.Hour * 24),
			}
			_, err := d.Con.Exec(d.Ctx, "UPDATE users SET cookie = $1 WHERE login IN ($2)", cookie, u.Login)
			if err != nil {
				d.Loger.Error("Error update cookie: ", err)
				return cookie, err
			}
			return cookie, err
		}
	} else {
		if err == nil {
			err = errors.New("invalid password")
		}
	}
	return cookie, err
}

func (d *Database) InitDB(ctx context.Context, DBURL string) error {
	DB.Ctx = ctx
	DB.DBURL = DBURL
	DB.Loger = logrus.New()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(2 * time.Second):
			var err error
			DB.Con, err = database.Connect(ctx, DBURL, d.Loger)
			if err != nil {
				DB.Loger.Error("Error conncet to DB: ", err)
				continue
			}
			break loop
		}
	}
	err := migrate.MigrateFromFS(ctx, DB.Con, &migrations.Migrations, DB.Loger)
	if err != nil {
		DB.Loger.Error("Error migration: ", err)
		return err
	}
	return nil
}

func (d *Database) CreateUser(u User) (cookie Cookie, err error) {
	hash, _ := app.HashPassword(u.Password)
	valhash := hmac.New(sha256.New, []byte(u.Login+hash))
	hh := fmt.Sprintf("%x", valhash.Sum(nil))
	switch d.Con {
	case nil:
		return cookie, err
	default:
		cookie := Cookie{
			Name:    "gophermart",
			Value:   hh,
			MaxAge:  86400,
			Expires: time.Now().Local().Add(time.Hour * 24),
		}
		_, err := d.Con.Exec(d.Ctx, "INSERT INTO users (login, password, cookie) VALUES($1,$2,$3)", u.Login, hash, cookie)
		if err != nil {
			d.Loger.Error("Error create user: ", err)
			return cookie, err
		}
		var userID int

		row := d.Con.QueryRow(d.Ctx, "SELECT user_id FROM users WHERE cookie->>'Value' = $1", cookie.Value)
		err = row.Scan(&userID)
		if err != nil {
			d.Loger.Error("Error scan row get user by cookie: ", err)
		}
		_, err = d.Con.Exec(d.Ctx, "INSERT INTO balance (user_id, balance) VALUES($1,$2)", userID, 0)
		return cookie, err
	}
}

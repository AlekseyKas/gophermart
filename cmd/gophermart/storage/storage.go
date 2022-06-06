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

type Order struct {
	Order   string
	Status  string
	Accrual float64
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

var DB Database
var IDB Storage
var Users User
var Orders Order

type Storage interface {
	InitDB(ctx context.Context, DBURL string) error
	CreateUser(u User) (cookie Cookie, err error)
	CheckCookie(c Cookie) (bool, error)
	// GetUser(u User) (string, error)
	AuthUser(u User) (Cookie, error)
	// ReconnectDB() error
	LoadOrder(number string, c Cookie) error
	UpdateOrder(order Order) error
}

// CREATE TABLE orders (
//   order_id INT NOT NULL GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//   user_id INT,
//   UNIQUE (user_id, number),
//   number VARCHAR ( 50 ) UNIQUE NOT NULL,
//   status VARCHAR (50) NOT NULL DEFAULT 'NEW',
//   accrual DOUBLE PRECISION DEFAULT 0,
//   uploaded_at TIMESTAMPTZ,
//   CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(user_id)
// );
func (d *Database) UpdateOrder(order Order) error {
	logrus.Info("[[[[[[[[[[[", order)
	_, err := d.Con.Exec(d.Ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3;", order.Status, order.Accrual, order.Order)
	if err != nil {
		logrus.Error("Error update accrual: ", err)
	}
	return err
}

func (d *Database) LoadOrder(number string, c Cookie) error {
	var login string
	var password string
	var cookie Cookie
	var userID int

	row := d.Con.QueryRow(d.Ctx, "SELECT * FROM users WHERE cookie->>'Value' = $1", c.Value)
	err := row.Scan(&userID, &login, &password, &cookie)
	if err != nil {
		d.Loger.Error("Error scan row get user by cookie: ", err)
		return err
	}
	_, err = d.Con.Exec(d.Ctx, "INSERT INTO orders (user_id, number, uploaded_at) VALUES($1,$2, $3)", userID, number, time.Now())
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
	logrus.Info("iIIIIIIIIIIIIIIIIIIIII", cookie)
	if cookie.Expires.After(time.Now()) {
		logrus.Info("iIIIIIIIIIIIIIIIIIIIIIaaaaaaaaaaaaaaaaa", cookie.Expires, time.Now())
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

// func (d *Database) GetUser(u User) (string, error) {
// 	var login string
// 	var password string
// 	var cookie Cookie
// 	row := d.Con.QueryRow(d.Ctx, "SELECT * FROM users WHERE login = $1", u.Login)
// 	err := row.Scan(&login, &password, &cookie)
// 	if err != nil {
// 		d.Loger.Error("Error scan row: ", err)
// 	}
// 	logrus.Info(")))))))))))))))))))", cookie.Expires.After(time.Now()))
// 	valhash := hmac.New(sha256.New, []byte(login+password))
// 	hh := fmt.Sprintf("%x", valhash.Sum(nil))
// 	return hh, nil
// }

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
		return cookie, err
	}
}

// func (d *Database) ReconnectDB() error {
// 	var err error
// 	for i := 0; i < 5; i++ {
// 		select {
// 		case <-d.Ctx.Done():
// 			return nil
// 		case <-time.After(2 * time.Second):
// 			DB.Con, err = database.Connect(d.Ctx, d.DBURL, d.Loger)
// 			if err != nil {
// 				d.Loger.Error("Error conncet to DB: ", err)
// 			} else {
// 				return nil
// 			}
// 		}
// 	}
// 	return err
// }

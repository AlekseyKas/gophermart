package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/handlers"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/helpers"
)

// Возможные коды ответа:
//     200 — пользователь успешно зарегистрирован и аутентифицирован;
//     400 — неверный формат запроса;
//     409 — логин уже занят;
//     500 — внутренняя ошибка сервера.
func Test_register(t *testing.T) {

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		body        []byte
		method      string
		url         string
		contentType string
		want        want
	}{
		// TODO: Add test cases.
		{
			name:        "success register",
			body:        []byte(`{"login": "user1", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/register",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:        "register duplicated",
			body:        []byte(`{"login": "user1", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/register",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  409,
			},
		},
		{
			name:        "wrong type#1",
			body:        []byte(`{"loginn": "user1", "sword": "password"}`),
			method:      "POST",
			url:         "/api/user/register",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  400,
			},
		},
		{
			name:        "wrong type#2",
			body:        []byte(`{"login": "user1", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/register",
			contentType: "plain/txt",
			want: want{
				contentType: "application/json",
				statusCode:  400,
			},
		},
	}
	// wg := &sync.WaitGroup{}
	ctx := context.Background()
	r := chi.NewRouter()
	// b := handlers.NewArgs(r, wg, ctx)
	r.Route("/", handlers.Router)
	ts := httptest.NewServer(r)
	defer ts.Close()

	DBURL, id, _ := helpers.StartDB()

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, DBURL)
	logrus.Info(DBURL)

	defer helpers.StopDB(id)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			buff := bytes.NewBuffer(body)
			req, err := http.NewRequest(tt.method, ts.URL+tt.url, buff)
			req.Header.Set("Content-Type", tt.contentType)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)

			require.NoError(t, err)
			defer resp.Body.Close()
		})
	}
}

// Возможные коды ответа:

//     200 — пользователь успешно аутентифицирован;
//     400 — неверный формат запроса;
//     401 — неверная пара логин/пароль;
//     500 — внутренняя ошибка сервера.
func Test_login(t *testing.T) {

	User := storage.User{
		Login:    "user1",
		Password: "password",
	}
	IPAddr := "127.0.0.1"

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		body        []byte
		method      string
		url         string
		contentType string
		want        want
	}{
		// TODO: Add test cases.
		{
			name:        "success login",
			body:        []byte(`{"login": "user1", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/login",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:        "wrong type#1",
			body:        []byte(`{"login": "user11", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/login",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  401,
			},
		},
		{
			name:        "wrong type#2",
			body:        []byte(`{"login": "user1", "password": "password"}`),
			method:      "POST",
			url:         "/api/user/login",
			contentType: "plain/txt",
			want: want{
				contentType: "application/json",
				statusCode:  400,
			},
		},
	}
	// wg := &sync.WaitGroup{}
	ctx := context.Background()

	r := chi.NewRouter()
	// b := handlers.NewArgs(r, wg, ctx)
	r.Route("/", handlers.Router)
	ts := httptest.NewServer(r)
	defer ts.Close()

	DBURL, id, _ := helpers.StartDB()

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, DBURL)
	logrus.Info(DBURL)
	storage.DB.CreateUser(User, IPAddr)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			buff := bytes.NewBuffer(body)
			req, err := http.NewRequest(tt.method, ts.URL+tt.url, buff)
			req.Header.Set("Content-Type", tt.contentType)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)

			require.NoError(t, err)
			defer resp.Body.Close()
		})
	}
	helpers.StopDB(id)
	time.Sleep(time.Second * 2)
	body := []byte(`{"login": "user1", "password": "password"}`)
	buff := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", ts.URL+"/api/user/login", buff)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 500, res.StatusCode)
	defer res.Body.Close()

}

// Возможные коды ответа:

// 200 — номер заказа уже был загружен этим пользователем;
// 202 — новый номер заказа принят в обработку;
// 400 — неверный формат запроса;
// 401 — пользователь не аутентифицирован;
// 409 — номер заказа уже был загружен другим пользователем;
// 422 — неверный формат номера заказа;
// 500 — внутренняя ошибка сервера.
func Test_loadOrder(t *testing.T) {
	User := storage.User{
		Login:    "user1",
		Password: "password",
	}

	User2 := storage.User{
		Login:    "user2",
		Password: "password",
	}

	User3 := storage.User{
		Login:    "user3",
		Password: "password",
	}
	IPAddr := "127.0.0.1"

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		body        []byte
		method      string
		url         string
		contentType string
		want        want
	}{
		// TODO: Add test cases.
		{
			name:        "wrong format req",
			method:      "GET",
			url:         "/api/user/orders",
			contentType: "application/json",
			want: want{
				contentType: "text/plain",
				statusCode:  204,
			},
		},
		{
			name:        "success registred",
			body:        []byte("12345678903"),
			method:      "POST",
			url:         "/api/user/orders",
			contentType: "text/plain",
			want: want{
				contentType: "text/plain",
				statusCode:  202,
			},
		},
		{
			name:        "success registred again from same user",
			body:        []byte("12345678903"),
			method:      "POST",
			url:         "/api/user/orders",
			contentType: "text/plain",
			want: want{
				contentType: "text/plain",
				statusCode:  200,
			},
		},
		{
			name:        "wrong format order",
			body:        []byte("00as"),
			method:      "POST",
			url:         "/api/user/orders",
			contentType: "text/plain",
			want: want{
				contentType: "text/plain",
				statusCode:  422,
			},
		},
		{
			name:        "wrong format req",
			body:        []byte(`"{number: 12345678903}"`),
			method:      "POST",
			url:         "/api/user/orders",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  400,
			},
		},
		{
			name:        "###get orders",
			method:      "GET",
			url:         "/api/user/orders",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:        "###wrong balance",
			method:      "GET",
			url:         "/api/user/balance",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:        "###get withdraw",
			method:      "GET",
			url:         "/api/user/balance/withdrawals",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  204,
			},
		},
		{
			name:        "###get withdraw",
			method:      "POST",
			body:        []byte(`"{order: "2377225624", "sum": 121}`),
			url:         "/api/user/balance/withdraw",
			contentType: "application/json",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
	}
	r := chi.NewRouter()
	ctx := context.Background()

	r.Route("/", handlers.Router)
	ts := httptest.NewServer(r)
	defer ts.Close()

	cmd := exec.Command("../../accrual/accrual_linux_amd64", "-a", "0.0.0.0:8090")
	err := cmd.Start()
	require.NoError(t, err)
	defer func(p *os.Process) {
		err := p.Kill()
		time.Sleep(time.Second * 7)
		require.NoError(t, err)
	}(cmd.Process)

	DBURL, id, _ := helpers.StartDB()

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, DBURL)
	//first user
	cookie, _ := storage.DB.CreateUser(User, IPAddr)
	cookie2, _ := storage.DB.CreateUser(User2, IPAddr)
	cookie3, _ := storage.DB.CreateUser(User3, IPAddr)

	testaccrualrequests(t, "http://0.0.0.0:8090")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			buff := bytes.NewBuffer(body)
			req, err := http.NewRequest(tt.method, ts.URL+tt.url, buff)
			cookie := &http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			}
			req.AddCookie(cookie)

			req.Header.Set("Content-Type", tt.contentType)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			require.NoError(t, err)
			defer resp.Body.Close()
		})
	}

	//409
	// //second user
	body := []byte("12345678903")
	buff := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", ts.URL+"/api/user/orders", buff)
	require.NoError(t, err)
	cookieReq := &http.Cookie{
		Name:  cookie2.Name,
		Value: cookie2.Value,
	}
	req.AddCookie(cookieReq)
	req.Header.Set("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 409, res.StatusCode)
	defer res.Body.Close()

	//after stop DB 401
	body = []byte("12345678903")
	buff = bytes.NewBuffer(body)
	req, err = http.NewRequest("POST", ts.URL+"/api/user/orders", buff)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 401, res.StatusCode)
	defer res.Body.Close()
	//stop DB
	helpers.StopDB(id)
	time.Sleep(time.Second * 2)

	//after stop DB 500
	body = []byte("12345678903")
	buff = bytes.NewBuffer(body)
	req, err = http.NewRequest("POST", ts.URL+"/api/user/orders", buff)
	require.NoError(t, err)
	cookieReq = &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	req.AddCookie(cookieReq)
	req.Header.Set("Content-Type", "text/plain")
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 500, res.StatusCode)
	defer res.Body.Close()

	//after stop DB 500 GET
	req, err = http.NewRequest("GET", ts.URL+"/api/user/orders", nil)
	require.NoError(t, err)
	cookieReq = &http.Cookie{
		Name:  cookie3.Name,
		Value: cookie3.Value,
	}
	req.AddCookie(cookieReq)
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 500, res.StatusCode)
	defer res.Body.Close()

}

func testaccrualrequests(t *testing.T, address string) {
	urlGoods := address + "/api/goods"
	urlOrders := address + "/api/orders"
	bodyGoods := `{ "match": "LG", "reward": 5, "reward_type": "%" }`
	bodyOrders := `{ "order": "123455", "goods": [ { "description": "LG Monitor", "price": 50000.0 } ] }`
	reqGoods, err := http.NewRequest("POST", urlGoods, bytes.NewReader([]byte(bodyGoods)))
	require.NoError(t, err)
	reqGoods.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(reqGoods)
	require.NoError(t, err)
	logrus.Info("Good request")
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	reqOrders, err := http.NewRequest("POST", urlOrders, bytes.NewReader([]byte(bodyOrders)))
	require.NoError(t, err)
	reqOrders.Header.Add("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(reqOrders)
	require.NoError(t, err)
	logrus.Info("Good order request")
	defer response.Body.Close()
	require.Equal(t, http.StatusAccepted, response.StatusCode)
}

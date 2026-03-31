package routers

import (
	"bytes"
	"encoding/json"
	"github.com/cenkalti/backoff/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/poggerr/gophermart/internal/app"
	"github.com/poggerr/gophermart/internal/async"
	"github.com/poggerr/gophermart/internal/config"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func NewDefConf() *config.Config {
	conf := config.Config{
		ServAddr: ":8081",
		DB:       "host=localhost user=gophermart password=userpassword dbname=gophermart sslmode=disable",
		Accrual:  ":8080",
		Client:   nil,
		Backoff:  nil,
	}
	conf.Client = &http.Client{}
	conf.Backoff = backoff.NewExponentialBackOff()
	conf.Backoff.MaxElapsedTime = 10 * time.Second
	return &conf
}

func testRequestPost(t *testing.T, ts *httptest.Server, method,
	path string, data string) (*http.Response, string) {

	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer([]byte(data)))
	require.NoError(t, err)

	//req.AddCookie(cookie)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func testRequestJSON(t *testing.T, ts *httptest.Server, method, path string, data []byte) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(data))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

type UserReg struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

func TestHandlersPost(t *testing.T) {
	// This is an integration-style test and requires PostgreSQL.
	cfg := NewDefConf()
	if v := os.Getenv("DATABASE_URI"); v != "" {
		cfg.DB = v
	}
	if os.Getenv("SECRET_KEY") == "" {
		_ = os.Setenv("SECRET_KEY", "test-secret")
	}

	db, err := sqlx.Connect("postgres", cfg.DB)
	if err != nil {
		t.Skipf("postgres is not available: %v", err)
		return
	}
	defer db.Close()

	sugaredLogger := logger.Initialize()
	strg := storage.NewStorage(db, cfg)
	repo := async.NewRepo(strg)
	newApp := app.NewApp(cfg, strg, sugaredLogger, repo)

	ts := httptest.NewServer(Router(newApp))
	defer ts.Close()

	var testTable = []struct {
		name        string
		api         string
		method      string
		url         string
		contentType string
		status      int
		location    string
	}{
		{name: "reg", api: "/api/user/register", method: "POST", status: 200},
		{name: "reg", api: "/api/user/register", method: "POST", status: 409},
		{name: "log", api: "/api/user/login", method: "POST", status: 200},
		{name: "log_negative", api: "/api/user/login", method: "POST", status: 401},
		{name: "make_order_without_auth", api: "/api/user/orders", method: "POST", status: 401},
	}

	for _, v := range testTable {
		switch {
		case v.name == "reg":
			user := UserReg{
				Username: "poggerr15",
				Password: "qwerty123",
			}
			marshal, err := json.Marshal(user)
			if err != nil {
				sugaredLogger.Info(err)
			}
			resp, _ := testRequestJSON(t, ts, v.method, v.api, marshal)
			defer resp.Body.Close()
			assert.Equal(t, v.status, resp.StatusCode)
		case v.name == "log":
			user := UserReg{
				Username: "poggerr15",
				Password: "qwerty123",
			}
			marshal, err := json.Marshal(user)
			if err != nil {
				sugaredLogger.Info(err)
			}
			resp, _ := testRequestJSON(t, ts, v.method, v.api, marshal)
			defer resp.Body.Close()
			assert.Equal(t, v.status, resp.StatusCode)
		case v.name == "log_negative":
			user := UserReg{
				Username: "poggerr1555",
				Password: "qwerty1237777",
			}
			marshal, err := json.Marshal(user)
			if err != nil {
				sugaredLogger.Info(err)
			}
			resp, _ := testRequestJSON(t, ts, v.method, v.api, marshal)
			defer resp.Body.Close()
			assert.Equal(t, v.status, resp.StatusCode)
		case v.name == "make_order_without_auth":
			num := "4736745983266662"
			resp, _ := testRequestPost(t, ts, v.method, v.api, num)
			defer resp.Body.Close()
			assert.Equal(t, v.status, resp.StatusCode)
		}

	}

}

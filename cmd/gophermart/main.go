package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/handlers"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/config"
	"github.com/AlekseyKas/gophermart/internal/middlewarecustom"
	lokihook "github.com/akkuman/logrus-loki-hook"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	lokiHookConfig := &lokihook.Config{
		URL: "https://logsremoteloki:efnd9DG510YnZQUjMlgMYVIN@loki.duduh.ru/api/prom/push",
		Labels: map[string]string{
			"app": "lexa-gophermart",
		},
	}
	hook, err := lokihook.NewHook(lokiHookConfig)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.AddHook(hook)
	}

	logrus.Info(">>>>>>>>>>>>>", config.Arg.DatabaseURL, "<<<<<<<<<<<<<<<<<<<<<<")
	//
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = config.TerminateFlags()
	if err != nil {
		logrus.Error("Error setting args: ", err)
	}

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, config.Arg.DatabaseURL)
	// wg.Add(1)
	// go app.WaitSignals(cancel, wg)

	r := chi.NewRouter()
	// b := handlers.NewArgs(r, wg, ctx)
	// s := &http.Server{
	// 	Handler: r,
	// 	// Addr:    "127.0.0.1:8080",
	// 	Addr: config.Arg.Address,
	// }
	r.Route("/", handlers.Router)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		logrus.Info("llllllllllllll")
		err := http.ListenAndServe("127.0.0.1:8080", r)
		if err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}(wg)
	// http.ListenAndServe("127.0.0.1:8080", r)
	<-ctx.Done()
	logrus.Info("Stop http server!")
	wg.Wait()
}
func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewarecustom.CheckCookie)
}

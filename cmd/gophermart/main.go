package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	lokihook "github.com/akkuman/logrus-loki-hook"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/handlers"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/helpers"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/app"
	"github.com/AlekseyKas/gophermart/internal/config"
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

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	err = config.TerminateFlags()
	if err != nil {
		logrus.Error("Error setting args: ", err)
	}
	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, config.Arg.DatabaseURL)
	wg.Add(3)
	go app.WaitSignals(cancel, wg)

	r := chi.NewRouter()
	s := &http.Server{
		Handler: r,
		Addr:    config.Arg.Address,
	}
	r.Route("/", handlers.Router)
	go helpers.ControlStatus(wg, ctx)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}(wg)
	<-ctx.Done()
	ctxx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.Shutdown(ctxx)
	logrus.Info("Http server stop!")
	wg.Wait()

}

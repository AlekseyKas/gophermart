package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/handlers"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/app"
	"github.com/AlekseyKas/gophermart/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	err := config.TerminateFlags()
	if err != nil {
		logrus.Error("Error setting args: ", err)
	}
	// logrus.Info(">>>>>>>>>>>>>", config.Arg.Address, "<<<<<<<<<<<<<<<<<<<<<<")

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, config.Arg.DatabaseURL)
	wg.Add(2)
	go app.WaitSignals(cancel, wg)

	r := chi.NewRouter()
	b := handlers.NewArgs(r, wg, ctx)
	s := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
		// Addr:    config.Arg.Address,
	}
	r.Route("/", b.Router)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}(wg)

	<-ctx.Done()
	s.Shutdown(ctx)
	logrus.Info("Stop http server!")
	wg.Wait()
}

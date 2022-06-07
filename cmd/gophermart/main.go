package main

import (
	"context"
	"net/http"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/config"
	"github.com/AlekseyKas/gophermart/internal/middlewarecustom"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	// wg := &sync.WaitGroup{}
	ctx, _ := context.WithCancel(context.Background())
	err := config.TerminateFlags()
	if err != nil {
		logrus.Error("Error setting args: ", err)
	}
	// logrus.Info(">>>>>>>>>>>>>", config.Arg.Address, "<<<<<<<<<<<<<<<<<<<<<<")

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
	r.Route("/", Router)
	// wg.Add(1)
	// go func(wg *sync.WaitGroup) {
	// 	defer wg.Done()
	// 	err := http.ListenAndServe("127.0.0.1:8080", r)
	// 	if err != nil && err != http.ErrServerClosed {
	// 		logrus.Error(err)
	// 	}
	// }(wg)

	http.ListenAndServe(config.Arg.Address, r)
	// <-ctx.Done()
	// s.Shutdown(ctx)
	// logrus.Info("Stop http server!")
	// wg.Wait()
}
func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewarecustom.CheckCookie)
}

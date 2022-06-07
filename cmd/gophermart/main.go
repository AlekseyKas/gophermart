package main

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	// wg := &sync.WaitGroup{}
	// ctx, cancel := context.WithCancel(context.Background())
	// err := config.TerminateFlags()
	// if err != nil {
	// 	logrus.Error("Error setting args: ", err)
	// }
	// // logrus.Info(">>>>>>>>>>>>>", config.Arg.Address, "<<<<<<<<<<<<<<<<<<<<<<")

	// storage.IDB = &storage.DB
	// storage.IDB.InitDB(ctx, config.Arg.DatabaseURL)
	// wg.Add(2)
	// go app.WaitSignals(cancel, wg)

	// r := chi.NewRouter()
	// b := handlers.NewArgs(r, wg, ctx)
	// s := &http.Server{
	// 	Handler: r,
	// 	Addr:    config.Arg.Address,
	// }
	// r.Route("/", b.Router)

	// go func(wg *sync.WaitGroup) {
	// 	defer wg.Done()
	// 	err := s.ListenAndServe()
	// 	if err != nil && err != http.ErrServerClosed {
	// 		logrus.Error(err)
	// 	}
	// }(wg)

	// <-ctx.Done()
	// s.Shutdown(ctx)
	// logrus.Info("Stop http server!")
	// wg.Wait()
	wg := &sync.WaitGroup{}
	r := chi.NewRouter()
	r.Route("/", Router)
	wg.Add(1)
	go http.ListenAndServe("127.0.0.1:8080", r)
	wg.Wait()
}

func Router(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

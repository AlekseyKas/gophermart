package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func WaitSignals(cancel context.CancelFunc, wg *sync.WaitGroup) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		sig := <-terminate
		switch sig {
		case os.Interrupt:
			defer wg.Done()
			cancel()
			logrus.Info("Terminate signal OS!")
			return
		}
	}
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// func RunAccurual(ctx context.Context, wg *sync.WaitGroup, address string) {
// 	defer wg.Done()
// 	cmd := exec.Command("cmd/accrual/accrual_linux_amd64")
// 	err := cmd.Start()
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// 	defer func(p *os.Process) {
// 		err = p.Kill()
// 		if err != nil {
// 			logrus.Error(err)
// 		}
// 	}(cmd.Process)
// 	<-ctx.Done()
// }

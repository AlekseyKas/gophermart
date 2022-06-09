package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/database"
	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/helpers"
)

func TestConnec(t *testing.T) {

	Loger := logrus.New()
	DBURL, id, _ := helpers.StartDB()
	ctx := context.Background()
	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, DBURL)
	t.Run("Valid and invalid connections", func(t *testing.T) {
		gotConPool, err := database.Connect(ctx, DBURL, Loger)
		require.NoError(t, err)
		require.NotEqual(t, gotConPool, nil)
		helpers.StopDB(id)
		time.Sleep(time.Second * 5)

		gotConPool2, err2 := database.Connect(ctx, DBURL, Loger)
		require.Error(t, err2)
		if gotConPool2 != nil {
			t.Errorf("Connect must be nil!")
		}
	})
}

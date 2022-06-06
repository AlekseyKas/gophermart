package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/gophermart/cmd/gophermart/storage"
	"github.com/AlekseyKas/gophermart/internal/helpers"
)

func TestDatabase_InitDB(t *testing.T) {

	t.Run("Init database", func(t *testing.T) {
		ctx := context.Background()
		DBURL, id, _ := helpers.StartDB()
		defer helpers.StopDB(id)

		storage.IDB = &storage.DB
		err := storage.IDB.InitDB(ctx, DBURL)
		require.NoError(t, err)

	})

}

func TestDatabase_User(t *testing.T) {
	type args struct {
		u storage.User
	}

	tests := []struct {
		name string
		args args
		want storage.Cookie
	}{
		{
			name: "first",
			args: args{storage.User{
				Login:    "user",
				Password: "password",
			}},
			want: storage.Cookie{
				Name:   "gophermart",
				Value:  "xxx",
				MaxAge: 864000,
			},
		},
		{
			name: "second",
			args: args{storage.User{
				Login:    "user2",
				Password: "password",
			}},
			want: storage.Cookie{
				Name:   "gophermart",
				Value:  "xxxxx",
				MaxAge: 99999,
			},
		},
		// TODO: Add test cases.
	}
	ctx := context.Background()
	DBURL, id, _ := helpers.StartDB()
	defer helpers.StopDB(id)

	storage.IDB = &storage.DB
	storage.IDB.InitDB(ctx, DBURL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie, err := storage.DB.CreateUser(tt.args.u)
			require.NoError(t, err)
			require.NotEmpty(t, cookie.Value)
			require.NotEmpty(t, cookie.MaxAge)
			require.NotEqual(t, cookie.MaxAge, tt.want.MaxAge)
			require.Equal(t, cookie.Name, tt.want.Name)
			// s, errg := storage.DB.GetUser(tt.args.u)
			// require.NoError(t, errg)
			// require.Equal(t, s, cookie.Value)

		})
	}
}

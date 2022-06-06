package app_test

import (
	"testing"

	"github.com/AlekseyKas/gophermart/internal/app"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "first",
			args: args{"test"},
		},
		{
			name: "second",
			args: args{"testtesttest"},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := app.HashPassword(tt.args.password)
			require.Equal(t, 60, len(s))
			require.NoError(t, err)
			result := app.CheckPasswordHash(tt.args.password, s)
			if !result {
				t.Errorf("Invalid hash %s password %s", s, tt.args.password)
			}
		})
	}
}

// func TestWaitSignals(t *testing.T) {
// 	wg := &sync.WaitGroup{}
// 	_, cancel := context.WithCancel(context.Background())
// 	type args struct {
// 		cancel context.CancelFunc
// 		wg     *sync.WaitGroup
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			name: "first",
// 			args: args{
// 				cancel: cancel,
// 				wg:     wg,
// 			},
// 		},
// 		// TODO: Add test cases.
// 	}
// 	// terminate := make(chan os.Signal, 1)
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			wg.Add(1)
// 			app.WaitSignals(tt.args.cancel, tt.args.wg)
// 			// sig := <-terminate
// 			time.Sleep(time.Second)
// 			wg.Done()
// 			// time.AfterFunc(1*time.Second, cancel)
// 		})
// 	}
// }

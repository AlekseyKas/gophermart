package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/gophermart/internal/app"
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

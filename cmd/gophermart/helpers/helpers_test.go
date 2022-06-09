package helpers

import (
	"context"
	"sync"
	"testing"
)

func TestControlStatus(t *testing.T) {
	type args struct {
		wg  *sync.WaitGroup
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ControlStatus(tt.args.wg, tt.args.ctx)
		})
	}
}

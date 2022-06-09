package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTerminateFlags(t *testing.T) {

	t.Run("Test terminate flag", func(t *testing.T) {
		err := TerminateFlags()
		require.NoError(t, err)
	})
}

func Test_loadConfig(t *testing.T) {

	t.Run("Test loading config", func(t *testing.T) {
		gotP, err := loadConfig()
		require.NoError(t, err)
		require.NotEqual(t, gotP, nil)
	})
}

package settings

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("should be ok without file", func(t *testing.T) {
		cfg := &Core{}

		v, err := New(cfg)
		require.NoError(t, err)
		require.Equal(t, v, Viper())
		require.IsType(t, viper.New(), v)
	})

	t.Run("should be ok with temp file", func(t *testing.T) {
		cfg := &Core{}

		tmpFile, err := os.CreateTemp(t.TempDir(), "example")
		require.NoError(t, err)

		cfg.File = tmpFile.Name()
		v, err := New(cfg)
		require.NoError(t, err)
		require.IsType(t, viper.New(), v)
	})

	t.Run("should fail", func(t *testing.T) {
		cfg := &Core{}

		cfg.File = "unknown file"
		v, err := New(cfg)
		require.Error(t, err)
		require.Nil(t, v)
	})
}

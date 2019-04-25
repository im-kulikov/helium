package settings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type providerType = func() *Core

func TestApp(t *testing.T) {

	t.Run("check provider", func(t *testing.T) {
		cfg := &Core{}

		provider := cfg.Provider()
		require.NotNil(t, provider)
		require.IsType(t, providerType(nil), provider.Constructor)
		appProvider := provider.Constructor.(providerType)
		require.Equal(t, cfg, appProvider())
	})

	t.Run("safe type", func(t *testing.T) {
		cfg := &Core{}

		cases := []string{"bad", "toml", "yml", "yaml"}
		for _, item := range cases {
			cfg.Type = item

			if item == "bad" {
				item = "yml"
			}

			require.Equal(t, item, cfg.SafeType())
		}
	})
}

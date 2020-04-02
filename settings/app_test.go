package settings

import (
	"testing"

	"go.uber.org/dig"

	"github.com/stretchr/testify/require"
)

type (
	coreProviderType = func() *Core
	diProviderType   = func() *dig.Container
)

func TestApp(t *testing.T) {
	t.Run("check di provider", func(t *testing.T) {
		di := dig.New()
		provider := DIProvider(di)
		require.NotNil(t, provider)
		require.IsType(t, diProviderType(nil), provider.Constructor)
		diProvider := provider.Constructor.(diProviderType)
		require.Equal(t, di, diProvider())
	})

	t.Run("check provider", func(t *testing.T) {
		cfg := &Core{}

		provider := cfg.Provider()
		require.NotNil(t, provider)
		require.IsType(t, coreProviderType(nil), provider.Constructor)
		appProvider := provider.Constructor.(coreProviderType)
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

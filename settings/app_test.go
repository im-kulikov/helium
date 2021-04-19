package settings

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
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
		diProvider, ok := provider.Constructor.(diProviderType)
		require.True(t, ok)
		require.Equal(t, di, diProvider())
	})

	t.Run("check provider", func(t *testing.T) {
		cfg := &Core{}

		provider := cfg.Provider()
		require.NotNil(t, provider)
		require.IsType(t, coreProviderType(nil), provider.Constructor)
		appProvider, ok := provider.Constructor.(coreProviderType)
		require.True(t, ok)
		require.Equal(t, cfg, appProvider())
	})

	t.Run("safe type", func(t *testing.T) {
		cfg := &Core{}

		cases := []string{"bad", "toml", "yml", "yaml"}
		for _, item := range cases {
			cfg.Type = item

			if item == "bad" {
				item = "yaml"
			}

			require.Equal(t, item, cfg.SafeType())
		}
	})
}

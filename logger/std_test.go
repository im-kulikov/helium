package logger

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStdLogger(t *testing.T) {
	std := NewStdLogger(zap.L())

	t.Run("not fatal calls should not panic", func(t *testing.T) {
		t.Run("print", func(t *testing.T) {
			require.NotPanics(t, func() { std.Print("panic no") })
		})

		t.Run("printf", func(t *testing.T) {
			require.NotPanics(t, func() { std.Printf("panic no") })
		})
	})

	t.Run("Fatal(f) should call os.Exit with 1 code", func(t *testing.T) {
		var exitCode int
		monkey.Patch(os.Exit, func(code int) { exitCode = code })
		defer monkey.Unpatch(os.Exit)

		t.Run("fatal", func(t *testing.T) {
			require.NotPanics(t, func() { std.Fatal("panic no") })
			require.Equal(t, 1, exitCode)
		})

		t.Run("fatalf", func(t *testing.T) {
			require.NotPanics(t, func() { std.Fatalf("panic no") })
			require.Equal(t, 1, exitCode)
		})
	})
}

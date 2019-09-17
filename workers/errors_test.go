package workers

import (
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    Error
		want string
	}{
		{name: "empty error"},
		{name: "custom error", e: Error("custom"), want: "custom"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

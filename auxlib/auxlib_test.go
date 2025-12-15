package auxlib

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/vm"
	"github.com/redis/go-redis/v9"
)

func TestService_Test(t *testing.T) {
	// We need a dummy storage for testing Test() since it only needs VM access really
	// Or we use RedisStorage if we assume redis is up.
	// The previous test used NewRedisScriptService which used redis.
	// Let's use RedisStorage for now as per previous behavior.

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	svc := NewService(NewRedisStorage(rdb, "icescript:"))

	tests := []struct {
		name     string
		content  string
		wantOut  string
		wantErr  bool
		errMatch string
	}{
		{
			name:    "print hello",
			content: `print("Hello World")`,
			wantOut: "Hello World",
			wantErr: false,
		},
		{
			name:    "math",
			content: `print(5 + 5)`,
			wantOut: "10",
			wantErr: false,
		},
		{
			name:     "syntax error",
			content:  `print(`,
			wantOut:  "",
			wantErr:  true,
			errMatch: "Parse errors",
		},
		{
			name:    "return value",
			content: `return 42`,
			wantOut: "Result: 42",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.Test(context.Background(), tt.content)
			if err != nil {
				t.Fatalf("Test() unexpected error: %v", err)
			}

			if tt.wantErr {
				if res.Error == "" {
					t.Errorf("Expected error, got none")
				} else if !strings.Contains(res.Error, tt.errMatch) {
					t.Errorf("Expected error match %q, got %q", tt.errMatch, res.Error)
				}
			} else {
				if res.Error != "" {
					t.Errorf("Unexpected error in result: %s", res.Error)
				}
				if !strings.Contains(res.Output, tt.wantOut) {
					t.Errorf("Output mismatch. want substring %q, got %q", tt.wantOut, res.Output)
				}
			}
		})
	}
}

func TestWithVmTestHarness(t *testing.T) {
	called := false
	// Dummy storage
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	storage := NewRedisStorage(rdb, "icescript:")

	svc := NewService(storage, WithVmTestHarness(func(ctx context.Context, v *vm.VM) error {
		called = true
		return nil
	}))

	svc.Test(context.Background(), `print("test")`)
	if !called {
		t.Errorf("Test harness was not called")
	}
}

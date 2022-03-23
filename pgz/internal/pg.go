package internal

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/ibrt/golang-fixtures/fixturez"

	"github.com/ibrt/golang-inject-pg/pgz"
)

var (
	_ fixturez.BeforeSuite = &ConfigHelper{}
)

// ConfigHelper is a test helper for *Config.
type ConfigHelper struct {
	// intentionally empty
}

// BeforeSuite implements fixturez.BeforeSuite.
func (h *ConfigHelper) BeforeSuite(ctx context.Context, t *testing.T) context.Context {
	t.Helper()

	return pgz.NewConfigSingletonInjector(&pgz.Config{
		PostgresURL: fmt.Sprintf(
			"postgres://postgres:password@localhost:%v/postgres",
			os.Getenv("POSTGRES_PORT")),
		EnableProxyMode:       true,
		ConnectTimeoutSeconds: 5,
	})(ctx)
}

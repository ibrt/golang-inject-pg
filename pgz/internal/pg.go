package internal

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/ibrt/golang-inject-pg/pgz"
)

// PGConfigHelper provides the *pgz.PGConfig for tests.
type PGConfigHelper struct {
	// intentionally empty
}

// BeforeSuite implements fixturez.BeforeSuite.
func (h *PGConfigHelper) BeforeSuite(ctx context.Context, _ *testing.T) context.Context {
	return pgz.NewConfigSingletonInjector(&pgz.PGConfig{
		PostgresURL:      fmt.Sprintf("postgres://postgres:password@localhost:%v/postgres", os.Getenv("POSTGRES_PORT")),
		EnableProxyMode:  true,
		ConnectTimeoutMS: 500,
	})(ctx)
}

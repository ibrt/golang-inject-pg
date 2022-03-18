package testpgz

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/ibrt/golang-fixtures/fixturez"

	"github.com/ibrt/golang-inject-pg/pgz"
)

var (
	_ fixturez.BeforeSuite = &Helper{}
	_ fixturez.AfterSuite  = &Helper{}
)

// Helper provides a test helper for logz using a real logger.
type Helper struct {
	origPostgresURL string
	dbName          string
	releaser        func()
}

// BeforeSuite implements fixturez.BeforeSuite.
func (f *Helper) BeforeSuite(ctx context.Context, _ *testing.T) context.Context {
	cfg := pgz.GetConfig(ctx)
	f.dbName = "test_" + uuid.Must(uuid.NewV4()).String()
	MustCreateDB(cfg.PostgresURL, f.dbName)
	f.origPostgresURL = cfg.PostgresURL
	cfg.PostgresURL = MustSelectDB(cfg.PostgresURL, f.dbName)

	injector, releaser := pgz.Initializer(ctx)
	f.releaser = releaser
	return injector(ctx)
}

// AfterSuite implements fixturez.AfterSuite.
func (f *Helper) AfterSuite(ctx context.Context, _ *testing.T) {
	f.releaser()
	f.releaser = nil

	cfg := pgz.GetConfig(ctx)
	cfg.PostgresURL = f.origPostgresURL
	MustDropDB(cfg.PostgresURL, f.dbName)
}

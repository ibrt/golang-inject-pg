package testpgz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ibrt/golang-errors/errorz"
	"github.com/ibrt/golang-fixtures/fixturez"

	"github.com/ibrt/golang-inject-pg/pgz"
)

var (
	_ fixturez.BeforeSuite = &Helper{}
	_ fixturez.AfterSuite  = &Helper{}
	_ fixturez.BeforeTest  = &Helper{}
)

// Helper provides a test helper for logz using a real logger.
type Helper struct {
	origPostgresURL string
	dbName          string
	releaser        func()
}

// BeforeSuite implements fixturez.BeforeSuite.
func (h *Helper) BeforeSuite(ctx context.Context, _ *testing.T) context.Context {
	cfg := pgz.GetConfig(ctx)
	h.dbName = "test_" + uuid.Must(uuid.NewV4()).String()
	MustCreateDB(cfg.PostgresURL, h.dbName)
	h.origPostgresURL = cfg.PostgresURL
	cfg.PostgresURL = MustSelectDB(cfg.PostgresURL, h.dbName)

	injector, releaser := pgz.Initializer(ctx)
	h.releaser = releaser
	return injector(ctx)
}

// AfterSuite implements fixturez.AfterSuite.
func (h *Helper) AfterSuite(ctx context.Context, _ *testing.T) {
	h.releaser()
	h.releaser = nil

	cfg := pgz.GetConfig(ctx)
	cfg.PostgresURL = h.origPostgresURL
	MustDropDB(cfg.PostgresURL, h.dbName)
}

// BeforeTest implements fixtures.BeforeTest.
func (h *Helper) BeforeTest(ctx context.Context, _ *testing.T) context.Context {
	h.resetNow(ctx)
	return ctx
}

// SetNow sets the "pg_now" function to return the given time.
// Note that this function rounds the time to microseconds for precision parity with Postgres, returns the rounded time.
func (h *Helper) SetNow(ctx context.Context, t time.Time) time.Time {
	t = t.Truncate(time.Microsecond)

	_, err := pgz.GetCtx(ctx).Exec(`
		CREATE OR REPLACE FUNCTION pg_now() RETURNS timestamptz AS
		$$ SELECT to_timestamp(` + fmt.Sprintf("%.6f", float64(t.UnixMicro())/1e6) + `); $$
		LANGUAGE SQL STABLE;
	`)
	errorz.MaybeMustWrap(err)
	return t
}

func (h *Helper) resetNow(ctx context.Context) {
	_, err := pgz.GetCtx(ctx).Exec(`
		CREATE OR REPLACE FUNCTION pg_now() RETURNS timestamptz AS
		$$ SELECT now(); $$
		LANGUAGE SQL STABLE;
	`)
	errorz.MaybeMustWrap(err)
}

package testpgz

import (
	"context"
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

// Helper is a test helper for PG.
type Helper struct {
	origPostgresURL string
	dbName          string
	releaser        func()
}

// BeforeSuite implements fixturez.BeforeSuite.
func (h *Helper) BeforeSuite(ctx context.Context, t *testing.T) context.Context {
	t.Helper()

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
func (h *Helper) AfterSuite(ctx context.Context, t *testing.T) {
	t.Helper()

	h.releaser()
	h.releaser = nil

	cfg := pgz.GetConfig(ctx)
	cfg.PostgresURL = h.origPostgresURL
	MustDropDB(cfg.PostgresURL, h.dbName)
}

// BeforeTest implements fixtures.BeforeTest.
func (h *Helper) BeforeTest(ctx context.Context, t *testing.T) context.Context {
	t.Helper()

	h.resetNow(ctx)
	return ctx
}

// SetNow sets the "pg_now" function to return the given time.
// Note that this function rounds the time to microseconds for precision parity with Postgres, returns the rounded time.
func (h *Helper) SetNow(ctx context.Context, t time.Time) time.Time {
	const layout = "2006-01-02 15:04:05.999999-07"
	t = t.Truncate(time.Microsecond).UTC()

	_, err := pgz.GetCtx(ctx).Exec(`
		CREATE OR REPLACE FUNCTION pg_now() RETURNS timestamptz AS
		$$ SELECT '` + t.Format(layout) + `'::timestamptz; $$
		LANGUAGE SQL STABLE;
	`)
	errorz.MaybeMustWrap(err, errorz.SkipPackage())
	return t
}

// GetNow calls the "pg_now" function.
func (h *Helper) GetNow(ctx context.Context) time.Time {
	row := pgz.GetCtx(ctx).QueryRow(`SELECT pg_now()`)
	errorz.MaybeMustWrap(row.Err(), errorz.SkipPackage())

	var now time.Time
	errorz.MaybeMustWrap(row.Scan(&now), errorz.SkipPackage())
	errorz.MaybeMustWrap(row.Err(), errorz.SkipPackage())

	return now.UTC()
}

func (h *Helper) resetNow(ctx context.Context) {
	_, err := pgz.GetCtx(ctx).Exec(`
		CREATE OR REPLACE FUNCTION pg_now() RETURNS timestamptz AS
		$$ SELECT now(); $$
		LANGUAGE SQL STABLE;
	`)
	errorz.MaybeMustWrap(err, errorz.SkipPackage())
}

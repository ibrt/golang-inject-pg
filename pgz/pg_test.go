package pgz_test

import (
	"context"
	"testing"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/ibrt/golang-fixtures/fixturez"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject-pg/pgz"
	"github.com/ibrt/golang-inject-pg/pgz/internal"
	"github.com/ibrt/golang-inject-pg/pgz/testpgz"
)

func TestDatabase(t *testing.T) {
	fixturez.RunSuite(t, &Suite{})
}

type Suite struct {
	*fixturez.DefaultConfigMixin
	PGConfig *internal.PGConfigHelper
	PG       *testpgz.Helper
}

func (s *Suite) TestPG(ctx context.Context, t *testing.T) {
	_, err := pgz.Get(ctx).ExecContext(ctx, `SELECT 1`)
	fixturez.RequireNoError(t, err)

	_, err = pgz.GetCtx(ctx).Exec(`SELECT 1`)
	fixturez.RequireNoError(t, err)

	rows, err := pgz.GetCtx(ctx).Query(`SELECT 1`)
	fixturez.RequireNoError(t, err)
	defer errorz.IgnoreClose(rows)
	var i int64
	require.True(t, rows.Next())
	fixturez.RequireNoError(t, rows.Scan(&i))
	require.False(t, rows.Next())
	require.Equal(t, int64(1), i)

	row := pgz.GetCtx(ctx).QueryRow(`SELECT 2`)
	fixturez.RequireNoError(t, row.Err())
	fixturez.RequireNoError(t, row.Scan(&i))
	fixturez.RequireNoError(t, row.Err())
	require.Equal(t, int64(2), i)
}

func TestFail(t *testing.T) {
	ctx := pgz.NewConfigSingletonInjector(&pgz.PGConfig{
		PostgresURL:      "postgres://postgres:password@localhost:3672/postgres",
		EnableProxyMode:  true,
		ConnectTimeoutMS: 500,
	})(context.Background())

	require.Panics(t, func() {
		injector, _ := pgz.Initializer(ctx)
		injector(ctx)
	})
}

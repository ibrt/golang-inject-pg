package testpgz_test

import (
	"context"
	"testing"
	"time"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/ibrt/golang-fixtures/fixturez"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject-pg/pgz"
	"github.com/ibrt/golang-inject-pg/pgz/internal"
	"github.com/ibrt/golang-inject-pg/pgz/testpgz"
)

func TestHelpers(t *testing.T) {
	fixturez.RunSuite(t, &Suite{})
}

type Suite struct {
	*fixturez.DefaultConfigMixin
	PGConfig *internal.PGConfigHelper
	PG       *testpgz.Helper
}

func (s *Suite) TestNow(ctx context.Context, t *testing.T) {
	now := s.PG.SetNow(ctx, time.Now().Add(-time.Hour))
	row := pgz.GetCtx(ctx).QueryRow(`SELECT pg_now()`)
	fixturez.RequireNoError(t, row.Err())
	var gNow time.Time
	fixturez.RequireNoError(t, row.Scan(&gNow))
	fixturez.RequireNoError(t, row.Err())
	require.Equal(t, gNow.UnixNano(), now.UnixNano())
}

func (s *Suite) TestCoverage(ctx context.Context, t *testing.T) {
	_, err := pgz.GetCtx(ctx).Exec(`
		CREATE OR REPLACE FUNCTION pg_increment(i integer) RETURNS integer AS 
		$$ BEGIN RETURN i + 1; END; $$
		LANGUAGE plpgsql;
	`)
	errorz.MaybeMustWrap(err)

	testpgz.ResetProfiler(ctx, "pg_increment")
	profile := testpgz.GetProfile(ctx, "pg_increment")
	require.Equal(t, 1.0, profile.BranchesTotal)
	require.Equal(t, 0.0, profile.StatementsTotal)

	row := pgz.GetCtx(ctx).QueryRow(`SELECT pg_increment(1)`)
	fixturez.RequireNoError(t, row.Err())
	var i int64
	fixturez.RequireNoError(t, row.Scan(&i))
	fixturez.RequireNoError(t, row.Err())
	require.Equal(t, int64(2), i)

	profile = testpgz.GetProfile(ctx, "pg_increment")
	profile.RequireFullCoverage(t)

	profile.PrettyPrint()
}

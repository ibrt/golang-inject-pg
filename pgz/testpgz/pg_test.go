package testpgz_test

import (
	"context"
	"testing"
	"time"

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

func (s *Suite) TestHelper(ctx context.Context, t *testing.T) {
	now := s.PG.SetNow(ctx, time.Now().Add(-time.Hour))
	row := pgz.GetCtx(ctx).QueryRow(`SELECT pg_now()`)
	fixturez.RequireNoError(t, row.Err())
	var gNow time.Time
	fixturez.RequireNoError(t, row.Scan(&gNow))
	fixturez.RequireNoError(t, row.Err())
	require.Equal(t, gNow.UnixNano(), now.UnixNano())
}

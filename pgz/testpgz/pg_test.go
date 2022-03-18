package testpgz_test

import (
	"context"
	"testing"

	"github.com/ibrt/golang-fixtures/fixturez"

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
	_, err := pgz.GetCtx(ctx).Exec(`SELECT 1`)
	fixturez.RequireNoError(t, err)
}

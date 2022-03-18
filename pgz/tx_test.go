package pgz_test

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/ibrt/golang-fixtures/fixturez"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/require"
	"go4.org/syncutil"

	"github.com/ibrt/golang-inject-pg/pgz"
)

func (s *Suite) TestTransaction_Retry(ctx context.Context, t *testing.T) {
	createTable(ctx, t)
	defer dropTable(ctx, t)

	gt := syncutil.NewGate(3)
	gp := &syncutil.Group{}

	for i := 0; i < 30; i++ {
		gt.Start()

		gp.Go(func() error {
			defer gt.Done()

			err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).Run(func(ctx context.Context) error {
				counter := readCounter(ctx, t)
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				_, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = $1 WHERE id = 0`, counter+1)
				return errorz.MaybeWrap(err)
			})
			return errorz.MaybeWrap(err)
		})
	}

	fixturez.RequireNoError(t, gp.Err())
	require.EqualValues(t, 30, readCounter(ctx, t))
}

func (s *Suite) TestTransaction_Reentrant(ctx context.Context, t *testing.T) {
	createTable(ctx, t)
	defer dropTable(ctx, t)

	err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).Run(func(ctx context.Context) error {
		if _, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`); err != nil {
			return errorz.Wrap(err)
		}

		err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).Run(func(ctx context.Context) error {
			_, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`)
			return errorz.MaybeWrap(err)
		})
		return errorz.MaybeWrap(err)
	})
	fixturez.RequireNoError(t, err)
	require.EqualValues(t, 2, readCounter(ctx, t))

	err = pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).Run(func(ctx context.Context) error {
		if _, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`); err != nil {
			return errorz.Wrap(err)
		}

		err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).SetAllowReentrant(false).Run(func(ctx context.Context) error {
			_, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`)
			return errorz.MaybeWrap(err)
		})
		return errorz.MaybeWrap(err)
	})
	require.EqualError(t, err, "unexpectedly nested transaction")
	require.EqualValues(t, 2, readCounter(ctx, t))
}

func (s *Suite) TestTransaction_ReadOnly(ctx context.Context, t *testing.T) {
	createTable(ctx, t)
	defer dropTable(ctx, t)

	err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).SetReadOnly(true).SetIsolationLevel(sql.LevelReadCommitted).Run(func(ctx context.Context) error {
		_, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`)
		return errorz.MaybeWrap(err)
	})
	require.Error(t, err)
	require.Equal(t, pgerrcode.ReadOnlySQLTransaction, errorz.Unwrap(err).(*pgconn.PgError).Code)
	require.EqualValues(t, 0, readCounter(ctx, t))
}

func (s *Suite) TestTransaction_Rollback(ctx context.Context, t *testing.T) {
	createTable(ctx, t)
	defer dropTable(ctx, t)

	err := pgz.NewTx(ctx).SetIsolationLevel(sql.LevelSerializable).Run(func(ctx context.Context) error {
		if _, err := pgz.GetCtx(ctx).Exec(`UPDATE test_transaction SET counter = counter + 1 WHERE id = 0`); err != nil {
			return errorz.Wrap(err)
		}

		_, err := pgz.GetCtx(ctx).Exec(`BAD`)
		return errorz.MaybeWrap(err)
	})
	require.Error(t, err)
	require.Equal(t, pgerrcode.SyntaxError, errorz.Unwrap(err).(*pgconn.PgError).Code)
	require.EqualValues(t, 0, readCounter(ctx, t))
}

func createTable(ctx context.Context, t *testing.T) {
	_, err := pgz.GetCtx(ctx).Exec(`CREATE TABLE test_transaction (id bigint NOT NULL PRIMARY KEY, counter bigint NOT NULL)`)
	fixturez.RequireNoError(t, err)

	_, err = pgz.GetCtx(ctx).Exec(`INSERT INTO test_transaction (id, counter) VALUES (0, 0)`)
	fixturez.RequireNoError(t, err)
}

func dropTable(ctx context.Context, t *testing.T) {
	_, err := pgz.GetCtx(ctx).Exec(`DROP TABLE test_transaction`)
	fixturez.RequireNoError(t, err)
}

func readCounter(ctx context.Context, t *testing.T) int64 {
	var counter int64
	row := pgz.GetCtx(ctx).QueryRow(`SELECT counter FROM test_transaction WHERE id = 0`)
	fixturez.RequireNoError(t, row.Err())
	fixturez.RequireNoError(t, row.Scan(&counter))
	return counter
}

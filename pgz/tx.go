package pgz

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

const (
	txMaxRetries = 10
)

// Tx describes a transaction.
type Tx struct {
	ctx            context.Context
	isolationLevel sql.IsolationLevel
	readOnly       bool
	allowReentrant bool
}

// NewTx initializes a new Tx.
func NewTx(ctx context.Context) *Tx {
	return &Tx{
		ctx:            ctx,
		isolationLevel: sql.LevelDefault,
		readOnly:       false,
		allowReentrant: true,
	}
}

// SetIsolationLevel sets the isolation level.
func (t *Tx) SetIsolationLevel(isolationLevel sql.IsolationLevel) *Tx {
	t.isolationLevel = isolationLevel
	return t
}

// SetReadOnly sets the read only flag.
func (t *Tx) SetReadOnly(readOnly bool) *Tx {
	t.readOnly = readOnly
	return t
}

// SetAllowReentrant sets the "allow reentrant" flag.
func (t *Tx) SetAllowReentrant(allowReentrant bool) *Tx {
	t.allowReentrant = allowReentrant
	return t
}

// Run runs the transaction.
func (t *Tx) Run(f func(ctx context.Context) error) error {
	if _, ok := t.ctx.Value(dbContextKey).(*sql.Tx); ok {
		if !t.allowReentrant {
			return errorz.Errorf("unexpectedly nested transaction", errorz.Skip())
		}
		return errorz.MaybeWrap(f(t.ctx), errorz.Skip())
	}

	for i := 0; i < txMaxRetries; i++ {
		err := t.runTxOnce(f)
		if err == nil {
			return nil
		}

		pgErr, ok := errorz.Unwrap(err).(*pgconn.PgError)
		if ok && pgErr.Code == pgerrcode.SerializationFailure && i < txMaxRetries-1 {
			time.Sleep(time.Duration(100 + (500*rand.Float64())*float64(time.Millisecond)))
			continue
		}

		return errorz.Wrap(err, errorz.Skip())
	}

	panic("unreachable")
}

func (t *Tx) runTxOnce(f func(ctx context.Context) error) error {
	tx, err := t.ctx.Value(dbContextKey).(*sql.DB).BeginTx(t.ctx, &sql.TxOptions{
		Isolation: t.isolationLevel,
		ReadOnly:  t.readOnly,
	})
	if err != nil {
		return errorz.Wrap(err, errorz.Skip())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := f(context.WithValue(t.ctx, dbContextKey, tx)); err != nil {
		return errorz.Wrap(err, errorz.Skip())
	}

	return errorz.MaybeWrap(tx.Commit(), errorz.Skip())
}

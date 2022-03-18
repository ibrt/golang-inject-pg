package pgz

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ibrt/golang-errors/errorz"
	"github.com/ibrt/golang-inject/injectz"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

type contextKey int

const (
	dbContextKey contextKey = iota
	pgConfigContextKey
)

var (
	validate = validator.New()
)

// PGConfig describes the configuration for the Postgres module.
type PGConfig struct {
	PostgresURL      string `json:"postgresUrl" validate:"required,url"`
	EnableProxyMode  bool   `json:"proxyMode"`
	ConnectTimeoutMS uint32 `json:"connectTimeoutMs"`
}

// NewConfigSingletonInjector always inject the given PGConfig.
func NewConfigSingletonInjector(cfg *PGConfig) injectz.Injector {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, pgConfigContextKey, cfg)
	}
}

// PG describes the Postgres module (a subset of *sql.DB).
type PG interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// ContextPG describes a PG with a cached context.
type ContextPG interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type contextPGImpl struct {
	ctx context.Context
	pg  PG
}

// Exec delegates to PG.ExecContext.
func (p *contextPGImpl) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.pg.ExecContext(p.ctx, query, args...)
}

// Query delegates to PG.QueryContext.
func (p *contextPGImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.pg.QueryContext(p.ctx, query, args...)
}

// QueryRow delegates to PG.QueryRowContext.
func (p *contextPGImpl) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.pg.QueryRowContext(p.ctx, query, args...)
}

// Initializer is an initializer for the Postgres module.
func Initializer(ctx context.Context) (injectz.Injector, injectz.Releaser) {
	cfg := ctx.Value(pgConfigContextKey).(*PGConfig)
	errorz.MaybeMustWrap(validate.Struct(cfg))

	pgxCfg, err := pgx.ParseConfig(cfg.PostgresURL)
	errorz.MaybeMustWrap(err)

	if cfg.EnableProxyMode {
		pgxCfg.BuildStatementCache = nil
		pgxCfg.PreferSimpleProtocol = true
	}

	pingCTX := ctx

	if timeout := time.Duration(cfg.ConnectTimeoutMS) * time.Millisecond; timeout > 0 {
		pgxCfg.ConnectTimeout = timeout

		var cancel context.CancelFunc
		pingCTX, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	db := stdlib.OpenDB(*pgxCfg)

	if err := db.PingContext(pingCTX); err != nil {
		errorz.IgnoreClose(db)
		errorz.MustWrap(err)
	}

	injector := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, dbContextKey, db)
	}

	releaser := func() {
		errorz.IgnoreClose(db)
	}

	return injector, releaser
}

// Get extracts the PG from context, panics if not found.
func Get(ctx context.Context) PG {
	return ctx.Value(dbContextKey).(PG)
}

// GetCtx extracts the PG from context and wraps it as ContextPG, panics if not found.
func GetCtx(ctx context.Context) ContextPG {
	return &contextPGImpl{
		ctx: ctx,
		pg:  Get(ctx),
	}
}

// GetConfig extracts the *PGConfig from context, panics if not found.
func GetConfig(ctx context.Context) *PGConfig {
	return ctx.Value(pgConfigContextKey).(*PGConfig)
}

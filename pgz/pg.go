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

// Config describes the configuration PG.
type Config struct {
	PostgresURL      string `json:"postgresUrl" validate:"required,url"`
	EnableProxyMode  bool   `json:"proxyMode"`
	ConnectTimeoutMS uint32 `json:"connectTimeoutMs"`
}

// NewConfigSingletonInjector always inject the given *Config.
func NewConfigSingletonInjector(cfg *Config) injectz.Injector {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, pgConfigContextKey, cfg)
	}
}

// GetConfig extracts the *Config from context, panics if not found.
func GetConfig(ctx context.Context) *Config {
	return ctx.Value(pgConfigContextKey).(*Config)
}

// PG describes the pg module (a subset of *sql.DB).
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

// Exec executes a query.
func (p *contextPGImpl) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.pg.ExecContext(p.ctx, query, args...)
}

// Query executes a query.
func (p *contextPGImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.pg.QueryContext(p.ctx, query, args...)
}

// QueryRow executes a query.
func (p *contextPGImpl) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.pg.QueryRowContext(p.ctx, query, args...)
}

// Initializer is a PG initializer.
func Initializer(ctx context.Context) (injectz.Injector, injectz.Releaser) {
	cfg := ctx.Value(pgConfigContextKey).(*Config)
	errorz.MaybeMustWrap(validate.Struct(cfg))

	pgxCfg, err := pgx.ParseConfig(cfg.PostgresURL)
	errorz.MaybeMustWrap(err, errorz.Skip())

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
		errorz.MustWrap(err, errorz.Skip())
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

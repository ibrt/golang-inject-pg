package testpgz

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ibrt/golang-errors/errorz"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

// MustOpen opens a Postgres DB and checks the connection, panics on error.
func MustOpen(pgURL string) *sql.DB {
	cfg, err := pgx.ParseConfig(pgURL)
	errorz.MaybeMustWrap(err)
	cfg.ConnectTimeout = 5 * time.Second
	db := stdlib.OpenDB(*cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errorz.MaybeMustWrap(db.PingContext(ctx))
	return db
}

// MustCreateDB creates (or recreates) a Postgres database, panics on error.
func MustCreateDB(pgURL, dbName string) {
	db := MustOpen(pgURL)
	defer errorz.IgnoreClose(db)
	escapedDBName := strings.ReplaceAll(dbName, `"`, `""`)

	_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%v"`, escapedDBName))
	errorz.MaybeMustWrap(err)

	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%v"`, escapedDBName))
	errorz.MaybeMustWrap(err)
}

// MustDropDB drops a Postgres database, panics on error.
func MustDropDB(pgURL, dbName string) {
	db := MustOpen(pgURL)
	defer errorz.IgnoreClose(db)

	_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%v"`, strings.ReplaceAll(dbName, `"`, `""`)))
	errorz.MaybeMustWrap(err)
}

// MustSelectDB returns a new pgURL with the given DB name, panics on error.
func MustSelectDB(pgURL, dbName string) string {
	parsedDBURL, err := url.Parse(pgURL)
	errorz.MaybeMustWrap(err)
	parsedDBURL.Path = dbName
	return parsedDBURL.String()
}

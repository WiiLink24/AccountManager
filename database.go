package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func initDatabase() {
	pool, err := pgxpool.New(ctx, config.Database)
	checkError(err)

	dbPool = pool
	ensureSchema()
}

func ensureSchema() {
	_, err := dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS push_subscriptions (
			endpoint   TEXT PRIMARY KEY,
			username   TEXT NOT NULL,
			p256dh     TEXT NOT NULL,
			auth       TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	checkError(err)

	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notification_preferences (
			username TEXT NOT NULL,
			category TEXT NOT NULL,
			enabled  BOOLEAN NOT NULL DEFAULT TRUE,
			PRIMARY KEY (username, category)
		);
	`)
	checkError(err)
}

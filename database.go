package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func initDatabase() {
	pool, err := pgxpool.New(ctx, config.Notifications.Database)
	checkError(err)

	dbPool = pool
}
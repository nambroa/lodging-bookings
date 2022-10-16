package dbrepo

import (
	"database/sql"
	"github.com/nambroa/lodging-bookings/internal/config"
	"github.com/nambroa/lodging-bookings/internal/repository"
)

// Repository pattern to abstract interactions with the DB.

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB // connection pool.
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{App: a, DB: conn}
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{App: a}
}

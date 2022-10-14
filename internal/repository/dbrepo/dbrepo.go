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

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{App: a, DB: conn}
}

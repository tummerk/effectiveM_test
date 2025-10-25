package persistent

import (
	"database/sql"
	"effectiveM_test/internal/domain/entity"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db}
}

func (r *PostgresRepo) Create(subscription entity.Subscription) error {
	
}

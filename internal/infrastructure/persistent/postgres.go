package persistent

import (
	"context"
	"database/sql"
	"effectiveMobile_test/internal/domain/entity"
	"errors"
	"fmt"
	"log/slog"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

type txKey struct{}

type executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *PostgresRepo) getExecutor(ctx context.Context) executor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return r.db
}

func (r *PostgresRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger(txCtx).Error("transaction rollback failed", slog.String("original_error", err.Error()), slog.String("rollback_error", rbErr.Error()))
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(txCtx)
	return err
}

func (r *PostgresRepo) Create(ctx context.Context, subscription entity.Subscription) (entity.Subscription, error) {
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at, updated_at`

	createdSub := subscription

	err := r.db.QueryRowContext(ctx, query,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserId,
		subscription.StartDate,
		subscription.EndDate,
	).Scan(&createdSub.ID, &createdSub.CreatedAt, &createdSub.UpdatedAt)

	if err != nil {
		logger(ctx).Error("Create Subscription Error", slog.String("error", err.Error()))
		return entity.Subscription{}, createSubscriptionError(err)
	}

	logger(ctx).Debug("Create Subscription Success", slog.Int("new_id", createdSub.ID))

	return createdSub, nil
}

func (r *PostgresRepo) GetById(ctx context.Context, id int) (entity.Subscription, error) {
	subscription := entity.Subscription{}
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
              FROM subscriptions WHERE id=$1`

	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserId,
		&subscription.StartDate,
		&subscription.EndDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Subscription{}, sql.ErrNoRows
	}
	if err != nil {
		logger(ctx).Error("GetById Subscription Error", slog.String("error", err.Error()))
		return entity.Subscription{}, getByIdError(err)
	}
	return subscription, nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id=$1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger(ctx).Error("Delete Subscription Error", slog.String("error", err.Error()))
		return deleteSubscriptionError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger(ctx).Warn("Could not get rows affected after delete", slog.String("error", err.Error()))
		return nil
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepo) GetAll(ctx context.Context, filter entity.ListFilter) ([]entity.Subscription, error) {
	query := `
		SELECT
			id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM
			subscriptions
		WHERE
			($1::uuid IS NULL OR user_id = $1) AND 
			($2::text IS NULL OR service_name = $2)
		ORDER BY
			created_at DESC;`
	rows, err := r.db.QueryContext(ctx, query, filter.UserId, filter.ServiceName)
	if err != nil {
		logger(ctx).Error("Error executing GetAll query", "error", err.Error())
		return nil, fmt.Errorf("failed to execute get all query: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]entity.Subscription, 0)
	for rows.Next() {
		var sub entity.Subscription
		err = rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserId,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			logger(ctx).Error("Error scanning subscription row", slog.String("error", err.Error()))
			return nil, scanningRowError(err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		logger(ctx).Error("Error during rows iteration", slog.String("error", err.Error()))
		return nil, iterationRowsError(err)
	}

	return subscriptions, nil
}

func (r *PostgresRepo) GetTotalCost(
	ctx context.Context, filter entity.CostFilter,
) (int, error) {
	query := `SELECT
				COALESCE(SUM(s.price), 0)::int AS total_spent
			FROM
				subscriptions s
				JOIN LATERAL generate_series(
					s.start_date,
					COALESCE(s.end_date, $2::DATE),
					'1 month'::interval
				) AS monthly_payment_date ON true
			WHERE
				monthly_payment_date >= $1::DATE
			AND monthly_payment_date <= $2::DATE
			AND ($3::uuid IS NULL OR s.user_id = $3)
            AND ($4::text IS NULL OR s.service_name = $4);`

	var cost int
	err := r.db.QueryRowContext(ctx, query,
		filter.DateStart, filter.DateEnd, filter.UserId, filter.ServiceName).Scan(&cost)
	if err != nil {
		logger(ctx).Error("Error calculating total cost", slog.String("error", err.Error()))
		return 0, calculatingTotalCostError(err)
	}

	return cost, nil
}

func (r *PostgresRepo) Update(ctx context.Context, subscription entity.Subscription) (entity.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET
			service_name = $1,
			price        = $2,
			user_id      = $3,
			start_date   = $4,
			end_date     = $5,
			updated_at   = now()
		WHERE id = $6
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at;
	`

	updated := entity.Subscription{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserId,
		subscription.StartDate,
		subscription.EndDate,
		subscription.ID,
	).Scan(
		&updated.ID,
		&updated.ServiceName,
		&updated.Price,
		&updated.UserId,
		&updated.StartDate,
		&updated.EndDate,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Subscription{}, sql.ErrNoRows
	}
	if err != nil {
		logger(ctx).Error("Update Subscription Error", slog.String("error", err.Error()))
		return entity.Subscription{}, updateSubscriptionError(err)
	}

	logger(ctx).Debug("Update Subscription Success", slog.Int("id", updated.ID))
	return updated, nil
}

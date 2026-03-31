package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"time"
)

func (strg *Storage) CreateWithdraw(userID *uuid.UUID, withdraw *models.Withdraw) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id := uuid.New()
	t := time.Now()
	t.Format(time.RFC3339)

	_, err := strg.DB.ExecContext(
		ctx,
		"INSERT INTO withdrawals (id, order_number, order_user, sum, processed_at) VALUES ($1, $2, $3, $4, $5)",
		id, withdraw.OrderNumber, userID, withdraw.Sum, t)
	if err != nil {
		logger.Initialize().Info(err)
		return err
	}
	return nil
}

func (strg *Storage) TakeUserWithdrawals(userID *uuid.UUID) (*models.Withdrawals, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := strg.DB.QueryContext(ctx, "SELECT * FROM withdrawals WHERE order_user=$1 ORDER BY processed_at", userID)
	if err != nil {
		logger.Initialize().Info(err)
		return nil, err
	}

	withdrawals := make(models.Withdrawals, 0)
	for rows.Next() {
		var withdraw models.Withdraw
		var id uuid.UUID
		var orderUser uuid.UUID
		if err = rows.Scan(&id, &withdraw.OrderNumber, &orderUser, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			logger.Initialize().Info(err)
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}

	if err = rows.Err(); err != nil {
		logger.Initialize().Info(err)
		return nil, err
	}
	return &withdrawals, nil
}

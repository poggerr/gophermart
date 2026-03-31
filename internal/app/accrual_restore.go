package app

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"time"
)

func (a *App) AccrualRestore() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := a.strg.DB.QueryContext(ctx, "SELECT * FROM orders WHERE status=$1 OR status=$2", "NEW", "PROCESSING")
	if err != nil {
		logger.Initialize().Info(err)
		return
	}
	for rows.Next() {
		var order models.UserOrder
		var id uuid.UUID
		var orderUser uuid.UUID
		var accrual sql.NullFloat64
		if err = rows.Scan(&id, &order.Number, &orderUser, &order.UploadedAt, &accrual, &order.Status); err != nil {
			logger.Initialize().Info(err)
		}
		if accrual.Valid {
			v := float32(accrual.Float64)
			order.Accrual = &v
		}
		a.repo.SendToChan(order.Number, &orderUser, a.cfg.Accrual)
	}

	if err = rows.Err(); err != nil {
		logger.Initialize().Info(err)
	}
}

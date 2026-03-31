package storage

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/accrualservice"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"time"
)

func (strg *Storage) TakeOrderByUser(orderNumber string) (*uuid.UUID, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user uuid.UUID

	ans := strg.DB.QueryRowContext(ctx, "SELECT order_user FROM orders WHERE order_number=$1", orderNumber)
	errScan := ans.Scan(&user)
	if errScan != nil {
		logger.Initialize().Info(errScan)
		return nil, false
	}
	return &user, true
}

func (strg *Storage) TakeUserOrders(userID *uuid.UUID) (*models.Orders, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := strg.DB.QueryContext(ctx, "SELECT * FROM orders WHERE order_user=$1 ORDER BY uploaded_at", userID)
	if err != nil {
		logger.Initialize().Info(err)
		return nil, err
	}

	orders := make(models.Orders, 0)
	for rows.Next() {
		var order models.UserOrder
		var id uuid.UUID
		var orderUser uuid.UUID
		var accrual sql.NullFloat64
		if err = rows.Scan(&id, &order.Number, &orderUser, &order.UploadedAt, &accrual, &order.Status); err != nil {
			logger.Initialize().Info(err)
			return nil, err
		}
		if accrual.Valid {
			v := float32(accrual.Float64)
			order.Accrual = &v
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		logger.Initialize().Info(err)
		return nil, err
	}
	return &orders, nil
}

type SaveOrd struct {
	OrderNum   string
	User       *uuid.UUID
	AccrualURL string
}

func (strg *Storage) SaveOrder(orderNumber string, user *uuid.UUID) error {
	t := time.Now()
	t.Format(time.RFC3339)
	id := uuid.New()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := strg.DB.ExecContext(
		ctx,
		"INSERT INTO orders (id, order_number, order_user, uploaded_at, status, accrual_service) VALUES ($1, $2, $3, $4, $5, $6)",
		id, orderNumber, user, t, "NEW", nil)
	if err != nil {
		logger.Initialize().Info(err)
		return err
	}
	return nil
}

func (strg *Storage) UpdateOrder(order SaveOrd) (done bool, retryAfter time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	accrual, err := accrualservice.Accrual(order.OrderNum, order.AccrualURL, strg.cfg.Client)
	if err != nil {
		if ra, ok := err.(*accrualservice.RetryAfterError); ok {
			return false, ra.After
		}
		logger.Initialize().Info(err)
		return false, 1 * time.Second
	}

	// map accrual statuses to user-visible statuses
	mappedStatus := accrual.Status
	switch accrual.Status {
	case "REGISTERED", "PROCESSING":
		mappedStatus = "PROCESSING"
	case "PROCESSED":
		mappedStatus = "PROCESSED"
	case "INVALID":
		mappedStatus = "INVALID"
	}

	var accrualValue any
	if accrual.Accrual != nil {
		accrualValue = *accrual.Accrual
	} else {
		accrualValue = nil
	}

	_, err = strg.DB.ExecContext(
		ctx,
		"UPDATE orders SET accrual_service=$1, status=$2 WHERE order_number=$3", accrualValue, mappedStatus, order.OrderNum)
	if err != nil {
		logger.Initialize().Info(err)
		return false, 1 * time.Second
	}

	if accrual.Accrual != nil && mappedStatus == "PROCESSED" {
		balance, err := strg.TakeUserBalance(order.User)
		if err != nil {
			logger.Initialize().Info(err)
		} else {
			balance.Current += *accrual.Accrual
			err = strg.UpdateUserBalance(order.User, balance.Current)
			if err != nil {
				logger.Initialize().Info(err)
			}
		}
	}

	switch mappedStatus {
	case "INVALID", "PROCESSED":
		return true, 0
	default:
		return false, 1 * time.Second
	}
}

package async

import (
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/storage"
	"time"
)

type AccrualRepo struct {
	takeAccrualChan chan storage.SaveOrd
	repository      storage.Storage
}

func NewRepo(strg *storage.Storage) *AccrualRepo {
	return &AccrualRepo{
		takeAccrualChan: make(chan storage.SaveOrd, 10),
		repository:      *strg,
	}
}

func (r *AccrualRepo) SendToChan(orderNum string, user *uuid.UUID, accrualURL string) {
	r.takeAccrualChan <- storage.SaveOrd{
		OrderNum:   orderNum,
		User:       user,
		AccrualURL: accrualURL,
	}
}

func (r *AccrualRepo) WorkerAccrual() {
	for accrual := range r.takeAccrualChan {
		done, after := r.repository.UpdateOrder(accrual)
		if done {
			continue
		}
		if after <= 0 {
			after = 1 * time.Second
		}
		next := accrual
		time.AfterFunc(after, func() {
			r.SendToChan(next.OrderNum, next.User, next.AccrualURL)
		})
	}

}

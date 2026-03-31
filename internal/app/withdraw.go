package app

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/authorization"
	"github.com/poggerr/gophermart/internal/models"
	"github.com/poggerr/gophermart/internal/ordervalidation"
	"io"
	"net/http"
	"strings"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

func (a *App) checkBalance(withdraw *models.Withdraw, userID *uuid.UUID) error {
	balance, err := a.strg.TakeUserBalance(userID)
	if err != nil {
		a.sugaredLogger.Info(err)
		return err
	}

	if balance.Current < withdraw.Sum {
		return ErrInsufficientFunds
	}
	return nil
}

func buildWithdraw(req *http.Request) (*models.Withdraw, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var withdraw models.Withdraw

	err = json.Unmarshal(body, &withdraw)
	if err != nil {
		return nil, err
	}

	return &withdraw, err
}

func (a *App) Withdraw(res http.ResponseWriter, req *http.Request) {
	userID := authorization.FromContext(req.Context())

	withdraw, err := buildWithdraw(req)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	withdraw.OrderNumber = strings.TrimSpace(withdraw.OrderNumber)
	if withdraw.OrderNumber == "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if !ordervalidation.OrderValidation(withdraw.OrderNumber) {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	_, isStore := a.strg.TakeOrderByUser(withdraw.OrderNumber)
	if isStore {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = a.checkBalance(withdraw, userID)
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			res.WriteHeader(http.StatusPaymentRequired)
		} else {
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = a.strg.Debit(userID, withdraw.Sum)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.strg.CreateWithdraw(userID, withdraw)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

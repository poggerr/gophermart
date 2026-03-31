package app

import (
	"github.com/poggerr/gophermart/internal/authorization"
	"github.com/poggerr/gophermart/internal/ordervalidation"
	"io"
	"net/http"
	"strings"
)

func (a *App) UploadOrder(res http.ResponseWriter, req *http.Request) {
	userID := authorization.FromContext(req.Context())

	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	isValid := ordervalidation.OrderValidation(orderNumber)
	if !isValid {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	user, isUsed := a.strg.TakeOrderByUser(orderNumber)
	if isUsed {
		switch *user {
		case *userID:
			res.WriteHeader(http.StatusOK)
			return
		default:
			res.WriteHeader(http.StatusConflict)
			return
		}
	}

	err = a.strg.SaveOrder(orderNumber, userID)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.repo.SendToChan(orderNumber, userID, a.cfg.Accrual)

	res.WriteHeader(http.StatusAccepted)

}

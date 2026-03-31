package app

import (
	"encoding/json"
	"github.com/poggerr/gophermart/internal/authorization"
	"net/http"
)

func (a *App) ListUserOrders(res http.ResponseWriter, req *http.Request) {
	userID := authorization.FromContext(req.Context())

	orders, err := a.strg.TakeUserOrders(userID)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(*orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	marshal, err := json.Marshal(orders)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(marshal)

}

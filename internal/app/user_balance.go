package app

import (
	"encoding/json"
	"github.com/poggerr/gophermart/internal/authorization"
	"net/http"
)

func (a *App) UserBalance(res http.ResponseWriter, req *http.Request) {
	userID := authorization.FromContext(req.Context())

	balance, err := a.strg.TakeUserBalance(userID)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	marshal, err := json.Marshal(balance)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(marshal)
}

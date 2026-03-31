package app

import (
	"encoding/json"
	"github.com/poggerr/gophermart/internal/authorization"
	"net/http"
)

func (a *App) Withdrawals(res http.ResponseWriter, req *http.Request) {
	userID := authorization.FromContext(req.Context())

	withdrawals, err := a.strg.TakeUserWithdrawals(userID)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(*withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	marshal, err := json.Marshal(withdrawals)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(marshal)
}

package app

import (
	"encoding/json"
	"github.com/poggerr/gophermart/internal/authorization"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"io"
	"net/http"
	"time"
)

func (a *App) RegisterUser(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user models.User

	err = json.Unmarshal(body, &user)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser, _ := a.strg.GetUser(user.Username)
	if dbUser != nil {
		res.WriteHeader(http.StatusConflict)
		return
	}

	userID, err := authorization.RegisterUser(a.strg, &user)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtString, err := authorization.BuildJWTString(&userID)
	if err != nil {
		logger.Initialize().Info(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	cook := &http.Cookie{
		Name:    "session_token",
		Value:   jwtString,
		Path:    "/",
		Domain:  "localhost",
		Expires: time.Now().Add(20 * time.Minute),
	}

	http.SetCookie(res, cook)

	res.WriteHeader(http.StatusOK)

}

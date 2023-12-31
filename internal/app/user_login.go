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

func (a *App) UserLogin(res http.ResponseWriter, req *http.Request) {
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

	newUser, err := a.strg.GetUser(user.Username)
	if err != nil {
		a.sugaredLogger.Info(err)
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = authorization.CheckPass(&user, newUser)
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwtString, err := authorization.BuildJWTString(&newUser.ID)
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

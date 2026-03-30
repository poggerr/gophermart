package authorization

import (
	"context"
	"github.com/google/uuid"
	"github.com/poggerr/gophermart/internal/logger"
	"net/http"
)

func AuthMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("session_token")
		if err != nil {
			logger.Initialize().Info(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user := GetUserID(c.Value)
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ur := r.WithContext(NewContext(r.Context(), user))

		//ur := r.WithContext(context.WithValue(r.Context(), "user", userid))

		h.ServeHTTP(w, ur)
	}
	return http.HandlerFunc(fn)
}

type userID string

const ReqUserKey = userID("userKey")

func NewContext(ctx context.Context, user *uuid.UUID) context.Context {
	return context.WithValue(ctx, ReqUserKey, user)
}

func FromContext(ctx context.Context) *uuid.UUID {
	u := ctx.Value(ReqUserKey).(*uuid.UUID)
	return u
}

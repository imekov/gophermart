package middlewares

import (
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/vladimirimekov/gophermart/internal/storage"
	"net/http"
)

type UserCookies struct {
	Storage storage.Repositories
	UserKey interface{}
}

func (h UserCookies) CheckUserCookies(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, data, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil || jwt.Validate(token) != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, ok := data["userID"].(float64)
		if !ok {
			http.Error(w, "wrong data in cookie", http.StatusInternalServerError)
			return
		}

		if ok = h.Storage.IsUserExistByUserID(int(userID), r.Context()); !ok {
			http.Error(w, "user doesn't exist", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})

}

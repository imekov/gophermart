package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h Handler) userHandle(w http.ResponseWriter, r *http.Request, handleType string) {
	var userID int
	g := UserData{}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if g.Login == "" || g.Password == "" {
		http.Error(w, "missing user or password", http.StatusBadRequest)
		return
	}

	isUserExist := h.Storage.IsUserExistByLogin(g.Login, r.Context())

	switch handleType {
	case "register":
		if isUserExist {
			http.Error(w, "user with this login already exist", http.StatusConflict)
			return
		}
		userID, err = h.Storage.CreateUser(g.Login, g.Password, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "login":
		if !isUserExist {
			http.Error(w, "invalid username/password ", http.StatusUnauthorized)
			return
		}
		userID, err = h.Storage.LoginUser(g.Login, g.Password, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	default:
		http.Error(w, "url not found", http.StatusNotFound)
		return
	}

	_, token, err := h.TokenAuth.Encode(map[string]interface{}{"userID": userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		Name:     "jwt",
		Value:    token,
	})

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	h.userHandle(w, r, "register")
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	h.userHandle(w, r, "login")
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Name:     "jwt",
		Value:    "",
	})

}

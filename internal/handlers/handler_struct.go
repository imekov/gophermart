package handlers

import (
	"github.com/go-chi/jwtauth"
	"github.com/vladimirimekov/gophermart/internal/storage"
)

type Handler struct {
	Storage   storage.Repositories
	TokenAuth *jwtauth.JWTAuth
	UserKey   interface{}
}

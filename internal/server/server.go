package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/go-chi/jwtauth"
	"github.com/vladimirimekov/gophermart/internal/middlewares"
	"github.com/vladimirimekov/gophermart/internal/storage"

	"github.com/go-chi/chi/v5"

	"github.com/vladimirimekov/gophermart/internal/handlers"

	"github.com/go-chi/chi/v5/middleware"
)

type userIDtype string

const userKey userIDtype = "userid"

type Order struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Server struct {
	dbConnection *sql.DB
	repository   *storage.PostgreConnect
}

func ConnectionInitialization(DatabaseURI string) (s Server) {
	var err error

	for i := 1; i <= 5; i++ {
		s.dbConnection, err = sql.Open("postgres", DatabaseURI)
		if err == nil {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if err != nil {
		log.Fatalf("unable to connect to database %v\n", DatabaseURI)
	}

	s.repository = storage.GetNewConnection(s.dbConnection, DatabaseURI)

	return
}

func (s Server) StartChi(Secret []byte) *chi.Mux {

	tokenAuth := jwtauth.New("HS256", Secret, nil)

	h := handlers.Handler{
		Storage:   s.repository,
		TokenAuth: tokenAuth,
		UserKey:   userKey,
	}

	m := middlewares.UserCookies{Storage: h.Storage, UserKey: userKey}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middlewares.GZIPRead)
	r.Use(middlewares.GZIPWrite)

	r.Route("/api/user/", func(r chi.Router) {

		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Get("/logout", h.Logout)

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(m.CheckUserCookies)

			r.Get("/orders", h.GetUserOrders)
			r.Post("/orders", h.PostUserOrders)
			r.Get("/withdrawals", h.GetUserWithdrawals)

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.GetUserBalance)
				r.Post("/withdraw", h.PostBalanceWithdraw)
			})

		})
	})

	return r
}

func (s Server) CloseConnection() {
	s.dbConnection.Close()
}

func (s Server) ExchangeWithAccrualSystem(AccrualSystemAddress string, ctx context.Context) {

	restyClient := resty.New()
	restyClient.SetBaseURL(AccrualSystemAddress)
	restyClient.SetHeader("Content-Type", "application/json")

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
			newOrders, err := s.repository.GetAllNewOrders(ctx)
			if err != nil {
				log.Print(err)
			}

			for _, order := range newOrders {
				urlGet := fmt.Sprintf("/api/orders/%s", order)
				order := Order{}
				restyClient.R().SetResult(&order).Get(urlGet)
				err := s.repository.UpdateOrderInformation(ctx, order.Order, order.Status, order.Accrual)
				if err != nil {
					log.Print(err)
				}

			}
		}
	}
}

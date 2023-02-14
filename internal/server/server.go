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

	"github.com/vladimirimekov/gophermart/internal/config"
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

func GetServer(dbConnection *sql.DB) (string, *chi.Mux) {

	cfg := config.GetConfig()
	tokenAuth := jwtauth.New("HS256", cfg.Secret, nil)

	var err error

	for i := 1; i <= 5; i++ {
		dbConnection, err = sql.Open("postgres", cfg.DatabaseURI)
		if err == nil {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if err != nil {
		log.Fatalf("unable to connect to database %v\n", cfg.DatabaseURI)
	}

	conn := storage.GetNewConnection(dbConnection, cfg.DatabaseURI)

	h := handlers.Handler{
		Storage:   conn,
		TokenAuth: tokenAuth,
		UserKey:   userKey,
	}

	m := middlewares.UserCookies{Storage: h.Storage, UserKey: userKey}

	ctx := context.Background()
	go func() {
		restyClient := resty.New()
		restyClient.SetBaseURL(cfg.AccrualSystemAddress)
		restyClient.SetHeader("Content-Type", "application/json")

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
				newOrders, err := conn.GetAllNewOrders(ctx)
				if err != nil {
					log.Print(err)
				}

				for _, order := range newOrders {
					urlGet := fmt.Sprintf("/api/orders/%s", order)
					order := Order{}
					restyClient.R().SetResult(&order).Get(urlGet)
					//TODO: проверить полученный заказ, если статус REGISTERED то пропустить
					err := conn.UpdateOrderInformation(ctx, order.Order, order.Status, order.Accrual)
					if err != nil {
						log.Print(err)
					}

				}
			}
		}
	}()

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

	return cfg.RunAddress, r
}

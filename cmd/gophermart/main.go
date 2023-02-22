package main

import (
	"context"
	"log"
	"net/http"

	"github.com/vladimirimekov/gophermart/internal/config"
	"github.com/vladimirimekov/gophermart/internal/server"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.GetConfig()
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	connection := server.ConnectionInitialization(cfg.DatabaseURI)
	defer connection.CloseConnection()

	go connection.ExchangeWithAccrualSystem(cfg.AccrualSystemAddress, ctx)

	log.Fatal(http.ListenAndServe(cfg.RunAddress, connection.StartChi(cfg.Secret)))

}

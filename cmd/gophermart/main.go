package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/vladimirimekov/gophermart/internal/server"
)

func main() {
	var dbConnection *sql.DB
	defer dbConnection.Close()

	log.Fatal(http.ListenAndServe(server.GetServer(dbConnection)))
}

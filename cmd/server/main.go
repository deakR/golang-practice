package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"eg/air/internal/api" 
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "demo.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		panic(err)
	}
	fmt.Println("Database connected!")

	apiCfg := api.ApiConfig{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", apiCfg.HealthHandler)
	mux.HandleFunc("/users", api.AuthMiddleware(apiCfg.UserHandler))
	mux.HandleFunc("/webhooks/payment", apiCfg.PaymentHandler)

	fmt.Println("Server started listening at port 7900")
	http.ListenAndServe(":7900", mux)
}
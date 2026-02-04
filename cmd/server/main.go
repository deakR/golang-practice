package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"basics2/internal/api"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found (this is fine for production if env vars are set)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}
	
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "demo.db"
	}

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		panic("AUTH_TOKEN is required in .env!")
	}

	db, err := sql.Open("sqlite", dbURL)
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

	apiCfg := api.ApiConfig{
		DB:         db,
		AuthSecret: authToken,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", apiCfg.HealthHandler)
	mux.HandleFunc("/users", apiCfg.AuthMiddleware(apiCfg.UserHandler))
	mux.HandleFunc("/webhooks/payment", apiCfg.PaymentHandler)

	fmt.Println("Server started listening at port 7900")
	http.ListenAndServe(":7900", mux)
}
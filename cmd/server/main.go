package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"basics2/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", apiCfg.HealthHandler)
	r.Post("/webhooks/payment", apiCfg.PaymentHandler)

	r.Route("/users", func(r chi.Router) {
		r.Use(apiCfg.AuthMiddleware)
		r.Get("/", apiCfg.GetUsers)
		r.Post("/", apiCfg.CreateUser)
		r.Put("/{id}", apiCfg.UpdateUser)
		r.Delete("/{id}", apiCfg.DeleteUser)
	})

	fmt.Println("Server started listening at port 7900")
	http.ListenAndServe(":"+port, r)
}
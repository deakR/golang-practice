package api

import (
	"database/sql"
	"net/http"
)

type ApiConfig struct {
	DB *sql.DB
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WebhookReq struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
		Amount int `json:"amount"`
	} `json:"data"`
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "DELETE" || r.Method == "PUT" {
			token := r.Header.Get("Authorization")
			if token != "Bearer standard_access_token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
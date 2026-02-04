package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/go-chi/chi/v5"
)

func sendEmail(email string) {
	time.Sleep(5 * time.Second)
	fmt.Printf("📧 Email sent to %s\n", email)
}

func (api *ApiConfig) HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "I am alive!")
}

func (api *ApiConfig) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := api.DB.Query("SELECT id, name, email FROM users")
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			continue
		}
		users = append(users, u)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (api *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Error decoding json", 400)
		return
	}

	result, err := api.DB.Exec("INSERT INTO users (name, email) VALUES (?, ?)", newUser.Name, newUser.Email)
	if err != nil {
		http.Error(w, "Failed to create user", 500)
		return
	}

	id, _ := result.LastInsertId()
	newUser.ID = int(id)
	go sendEmail(newUser.Email)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
	fmt.Printf("Created user: %+v\n", newUser)
}

func (api *ApiConfig) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	_, err = api.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %d deleted", id)
}

func (api *ApiConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid Body", 400)
		return
	}

	_, err = api.DB.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", updatedUser.Name, updatedUser.Email, id)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	updatedUser.ID = id
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func (api *ApiConfig) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var hook WebhookReq
		if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
			http.Error(w, "Error decoding json", 400)
			return
		}
		if hook.Event == "payment_success" {
			fmt.Printf("💰 Payment of $%d received from User %d\n", hook.Data.Amount, hook.Data.UserID)
			w.WriteHeader(http.StatusOK)
		} else {
			fmt.Printf("Ignored event: %s\n", hook.Event)
		}
	}
}
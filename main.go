package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"database/sql"
	_ "modernc.org/sqlite"	
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "demo.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err!=nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
    if err != nil {
        panic(err)
    }

    fmt.Println("Database connected!")

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/users", authMiddleware(userHandler))
	mux.HandleFunc("/webhooks/payment", paymentHandler)
	fmt.Println("Server started listening at port 7900")
	http.ListenAndServe(":7900", mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "I am alive!")
}

// homeHandler serves as the API documentation homepage, listing all available endpoints
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<html>
		<head><title>API Documentation</title></head>
		<body>
			<h1>Welcome to the API</h1>
			<h2>Available Endpoints:</h2>
			<ul>
				<li><strong>GET /health</strong> - Check API status</li>
				<li><strong>GET /users</strong> - List all users</li>
				<li><strong>POST /users</strong> - Create a new user (Requires Admin Token)</li>
				<li><strong>POST /webhooks/payment</strong> - Handle payment events</li>
			</ul>
		</body>
		</html>
	`)
}

// authMiddleware wraps handlers to check for valid authorization tokens on POST requests
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "DELETE" || r.Method == "PUT" {
			token := r.Header.Get("Authorization")

			if token != "Bearer standard_access_token" {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Unauthorized")
				return
			}
		}
		next(w, r)
	}
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

// userHandler handles GET and POST requests for the /users endpoint
// GET returns all users, POST creates a new user (requires authorization)
func userHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		
		rows, err := db.Query("SELECT id, name, email FROM users")
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


	} else if r.Method == "POST" {
		var newUser User
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			http.Error(w, "Error decoding json", 400)
			return
		}

		result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", newUser.Name, newUser.Email)
		if err != nil {
			http.Error(w, "Failed to create user", 500)
			return
		}
		
		id, _ := result.LastInsertId()
		newUser.ID = int(id)
		
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newUser)
		fmt.Printf("Created user: %+v\n", newUser)


	} else if r.Method == "DELETE" {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", 400)
			return
		}

		_, err = db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User %d deleted", id)

	} else if r.Method == "PUT" {
		idStr := r.URL.Query().Get("id")
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

		_, err = db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", updatedUser.Name, updatedUser.Email, id)
		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		updatedUser.ID = id
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedUser)
	}
}

// paymentHandler processes incoming webhook requests for payment events
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var hook WebhookReq
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&hook)
		if err != nil {
			fmt.Fprint(w, "Error decoding json")
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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
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
		if r.Method == "POST" {
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

var users = []User{
	{ID: 1, Name: "Leon S. Kennedy", Email: "123@gmail.com"},
	{ID: 2, Name: "Ada Wong", Email: "321@gmail.com"},
}

// userHandler handles GET and POST requests for the /users endpoint
// GET returns all users, POST creates a new user (requires authorization)
func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	} else if r.Method == "POST" {
		var newUser User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newUser)
		if err != nil {
			fmt.Fprint(w, "Error decoding json")
			return
		}

		newUser.ID = len(users) + 1
		users = append(users, newUser)
		w.WriteHeader(http.StatusCreated)
		fmt.Printf("created %+v\n", newUser)
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

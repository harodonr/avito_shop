package main

import (
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

var jwtKey = []byte("secret-key") // Для демонстрации, в production храните в защите

func main() {
	// Инициализация in-memory хранилища пользователей
	InitStore()

	// Маршруты
	r := mux.NewRouter()

	// Аутентификация
	r.HandleFunc("/auth", AuthHandler).Methods("POST")

	// Защищённые маршруты
	api := r.PathPrefix("/me").Subrouter()
	api.Use(JWTMiddleware)
	api.HandleFunc("/merch", GetUserMerchHandler).Methods("GET")
	api.HandleFunc("/buy", BuyMerchHandler).Methods("POST")
	api.HandleFunc("/transfer", TransferHandler).Methods("POST")
	api.HandleFunc("/transactions", GetTransactionsHandler).Methods("GET")

	log.Println("Запуск сервера на порту 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// Определение структуры для JWT Claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}


package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	// Подключение к PostgreSQL (в Docker Compose использует имя "postgres" как хост)
	connStr := "postgres://user:password@postgres:5432/merch_shop?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
}

var jwtKey = []byte("secret-key") // Для демонстрации, в production храните в защите

func main() {
    // Инициализация маршрутов
    r := mux.NewRouter()

    // Маршрут аутентификации без проверки JWT
    r.HandleFunc("/api/auth", AuthHandler).Methods("POST")
    // Применяем JWTMiddleware ко всем маршрутам, которые требуют авторизации
    api := r.PathPrefix("/api").Subrouter()
    api.Use(JWTMiddleware)
    api.HandleFunc("/info", InfoHandler).Methods("GET")
    api.HandleFunc("/sendCoin", SendCoinHandler).Methods("POST")
    api.HandleFunc("/buy/{item}", BuyMerchHandler).Methods("GET")

    // Настроим маршруты для защищённых функций
    apiMe := r.PathPrefix("/me").Subrouter()
    apiMe.Use(JWTMiddleware)
    apiMe.HandleFunc("/merch", GetUserMerchHandler).Methods("GET")
    apiMe.HandleFunc("/transfer", TransferHandler).Methods("POST")
    apiMe.HandleFunc("/transactions", GetTransactionsHandler).Methods("GET")

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


package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gorilla/mux"
)

func getTokenForUser(t *testing.T, username string) string {
	// Создаем запрос для авторизации
	body := map[string]string{"username": username}
	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/auth", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AuthHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Unexpected status code: %v", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	token, exists := resp["token"]
	if !exists {
		t.Fatal("Token not returned")
	}
	return token
}

func TestBuyMerch(t *testing.T) {
	// Инициализация маршрутизатора
	r := mux.NewRouter()
	r.HandleFunc("/auth", AuthHandler).Methods("POST")
	api := r.PathPrefix("/me").Subrouter()
	api.Use(JWTMiddleware)
	api.HandleFunc("/buy", BuyMerchHandler).Methods("POST")

	// Авторизуем пользователя
	token := getTokenForUser(t, "test_user_buy")

	// Покупка существующего товара
	body := map[string]string{"item": "t-shirt"}
	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/me/buy", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}
}

func TestTransferCoins(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/auth", AuthHandler).Methods("POST")
	api := r.PathPrefix("/me").Subrouter()
	api.Use(JWTMiddleware)
	api.HandleFunc("/transfer", TransferHandler).Methods("POST")

	// Авторизуем отправителя
	senderToken := getTokenForUser(t, "sender")
	// Авторизуем получателя (автоматически создастся при передаче)
	_ = getTokenForUser(t, "recipient")

	// Перевод монеток
	body := map[string]interface{}{
		"to":    "recipient",
		"coins": 100,
	}
	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/me/transfer", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+senderToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	// Проверка баланса отправителя
	sender := GetUser("sender")
	if sender.Coins != 900 {
		t.Fatalf("Expected sender coins 900, got %d", sender.Coins)
	}

	// Проверка баланса получателя
	recipient := GetUser("recipient")
	if recipient.Coins != 1100 {
		t.Fatalf("Expected recipient coins 1100, got %d", recipient.Coins)
	}
}


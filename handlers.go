package main

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt"
)

// AuthHandler выполняет аутентификацию/регистрацию
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Получаем или создаём пользователя
	user := GetOrCreateUser(req.Username)

	// Создание JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Невозможно создать токен", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"token": tokenString}
	json.NewEncoder(w).Encode(resp)
}

// GetUserMerchHandler возвращает список купленных мерч товаров
func GetUserMerchHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	user := GetUser(username)
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.PurchasedMerch)
}

// BuyMerchHandler реализует покупку мерча
func BuyMerchHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	var req struct {
		Item string `json:"item"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Item == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Получаем информацию о товаре
	item, ok := MerchCatalog[req.Item]
	if !ok {
		http.Error(w, "Неизвестный товар", http.StatusBadRequest)
		return
	}

	user := GetUser(username)
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	if user.Coins < item.Price {
		http.Error(w, "Недостаточно монеток", http.StatusBadRequest)
		return
	}

	// Списание монет и внесение товара в список покупок
	_ = user.BuyMerch(item)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Покупка совершена", "item": item.Name})
}

// TransferHandler реализует передачу монеток между сотрудниками
func TransferHandler(w http.ResponseWriter, r *http.Request) {
	senderUsername := r.Context().Value("username").(string)
	var req struct {
		To    string `json:"to"`
		Coins int    `json:"coins"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.To == "" || req.Coins <= 0 {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	sender := GetUser(senderUsername)
	recipient := GetOrCreateUser(req.To) // Создаем пользователя, если не существует

	if sender.Coins < req.Coins {
		http.Error(w, "Недостаточно монеток для перевода", http.StatusBadRequest)
		return
	}

	err := sender.TransferCoins(recipient, req.Coins)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Перевод выполнен"})
}

// GetTransactionsHandler возвращает историю переводов (входящие и исходящие)
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	user := GetUser(username)
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"incoming": user.IncomingTransfers,
		"outgoing": user.OutgoingTransfers,
	}
	json.NewEncoder(w).Encode(resp)
}


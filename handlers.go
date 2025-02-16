package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

// AuthHandler выполняет аутентификацию/регистрацию и создание JWT токена.
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	// Получаем или создаём пользователя
	user, err := GetUserByUsername(req.Username)
	if err != nil {
		user, err = CreateUser(req.Username, req.Password) // Создаём нового пользователя
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
			return
		}
	}

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
		http.Error(w, "Не удалось создать токен", http.StatusInternalServerError)
		return
	}

	// Отправляем токен пользователю
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: tokenString})
}

// InfoHandler возвращает информацию о монетах, инвентаре и истории транзакций.
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Получаем информацию о монетах и инвентаре
	coins := user.Coins
	inventory := make([]InventoryItem, 0)
	coinHistory := CoinHistory{}

	// Получаем инвентарь (купленные товары)
	rows, err := db.Query(`
		SELECT m.name, COUNT(*) 
		FROM merchandise m
		JOIN purchases p ON m.id = p.merchandise_id
		WHERE p.user_id = $1
		GROUP BY m.name
	`, user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении инвентаря", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			http.Error(w, "Ошибка при сканировании инвентаря", http.StatusInternalServerError)
			return
		}
		inventory = append(inventory, item)
	}

	// Получаем историю монет
	rows, err = db.Query(`
		SELECT u.username, t.amount 
		FROM transactions t
		JOIN users u ON u.id = t.sender_id
		WHERE t.receiver_id = $1
	`, user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении истории монет", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var transfer TransferInfo
		if err := rows.Scan(&transfer.FromUser, &transfer.Amount); err != nil {
			http.Error(w, "Ошибка при сканировании истории монет", http.StatusInternalServerError)
			return
		}
		coinHistory.Received = append(coinHistory.Received, transfer)
	}

	// Получаем переводы монет, которые отправил пользователь
	rows, err = db.Query(`
		SELECT u.username, t.amount 
		FROM transactions t
		JOIN users u ON u.id = t.receiver_id
		WHERE t.sender_id = $1
	`, user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении истории монет", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var transfer TransferInfo
		if err := rows.Scan(&transfer.ToUser, &transfer.Amount); err != nil {
			http.Error(w, "Ошибка при сканировании истории монет", http.StatusInternalServerError)
			return
		}
		coinHistory.Sent = append(coinHistory.Sent, transfer)
	}

	// Формируем ответ
	infoResponse := InfoResponse{
		Coins:     coins,
		Inventory: inventory,
		CoinHistory: coinHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(infoResponse)
}

// SendCoinHandler выполняет перевод монет между пользователями.
func SendCoinHandler(w http.ResponseWriter, r *http.Request) {
	var req SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ToUser == "" || req.Amount <= 0 {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	senderUsername := r.Context().Value("username").(string)
	sender, err := GetUserByUsername(senderUsername)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	recipient, err := GetUserByUsername(req.ToUser)
	if err != nil {
		http.Error(w, "Получатель не найден", http.StatusNotFound)
		return
	}

	if sender.Coins < req.Amount {
		http.Error(w, "Недостаточно монет для перевода", http.StatusBadRequest)
		return
	}

	// Перевод монет
	err = sender.TransferCoins(recipient, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Перевод выполнен"})
}

// BuyMerchHandler выполняет покупку товара за монеты.
func BuyMerchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemName := vars["item"]

	// Получаем товар из базы
	item, err := GetMerchandiseByName(itemName)
	if err != nil {
		http.Error(w, "Товар не найден", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)
	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	if user.Coins < item.Price {
		http.Error(w, "Недостаточно монет для покупки", http.StatusBadRequest)
		return
	}

	// Покупка товара
	err = user.BuyMerch(item)
	if err != nil {
		http.Error(w, "Ошибка при покупке товара", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Покупка совершена", "item": item.Name})
}

// GetUserMerchHandler возвращает список купленных пользователем товаров.
func GetUserMerchHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Получаем список купленных товаров
	var purchasedMerch []Merchandise
	for _, item := range user.PurchasedMerch {
		purchasedMerch = append(purchasedMerch, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(purchasedMerch)
}

// TransferHandler выполняет перевод монет между пользователями.
func TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ToUser == "" || req.Amount <= 0 {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	senderUsername := r.Context().Value("username").(string)
	sender, err := GetUserByUsername(senderUsername)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	recipient, err := GetUserByUsername(req.ToUser)
	if err != nil {
		http.Error(w, "Получатель не найден", http.StatusNotFound)
		return
	}

	if sender.Coins < req.Amount {
		http.Error(w, "Недостаточно монет для перевода", http.StatusBadRequest)
		return
	}

	// Перевод монет
	err = sender.TransferCoins(recipient, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Перевод выполнен"})
}

// GetTransactionsHandler возвращает историю транзакций (входящие и исходящие).
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Получаем входящие переводы
	incomingTransfers := make([]TransferInfo, 0)
	rows, err := db.Query(`
		SELECT u.username, t.amount
		FROM transactions t
		JOIN users u ON u.id = t.sender_id
		WHERE t.receiver_id = $1
	`, user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении входящих переводов", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var transfer TransferInfo
		if err := rows.Scan(&transfer.FromUser, &transfer.Amount); err != nil {
			http.Error(w, "Ошибка при сканировании входящих переводов", http.StatusInternalServerError)
			return
		}
		incomingTransfers = append(incomingTransfers, transfer)
	}

	// Получаем исходящие переводы
	outgoingTransfers := make([]TransferInfo, 0)
	rows, err = db.Query(`
		SELECT u.username, t.amount
		FROM transactions t
		JOIN users u ON u.id = t.receiver_id
		WHERE t.sender_id = $1
	`, user.ID)
	if err != nil {
		http.Error(w, "Ошибка при получении исходящих переводов", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var transfer TransferInfo
		if err := rows.Scan(&transfer.ToUser, &transfer.Amount); err != nil {
			http.Error(w, "Ошибка при сканировании исходящих переводов", http.StatusInternalServerError)
			return
		}
		outgoingTransfers = append(outgoingTransfers, transfer)
	}

	// Формируем ответ
	transactionHistory := map[string]interface{}{
		"incoming": incomingTransfers,
		"outgoing": outgoingTransfers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactionHistory)
}


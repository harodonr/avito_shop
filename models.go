package main

import (
	"database/sql"
	"fmt"
	//"log"
)

// AuthRequest - структура запроса для аутентификации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse - структура ответа на аутентификацию
type AuthResponse struct {
	Token string `json:"token"`
}

// InventoryItem - структура для элемента инвентаря
type InventoryItem struct {
	Type     string `json:"type"`     // Тип предмета (например, "t-shirt")
	Quantity int    `json:"quantity"` // Количество предметов
}

// SendCoinRequest - структура для запроса перевода монет
type SendCoinRequest struct {
    ToUser string `json:"toUser"` // Имя получателя монет
    Amount int    `json:"amount"` // Количество монет
}


// CoinHistory - структура для истории монет
type CoinHistory struct {
	Received []TransferInfo `json:"received"` // Переводы монет, полученных пользователем
	Sent     []TransferInfo `json:"sent"`     // Переводы монет, отправленных пользователем
}

// TransferInfo - структура для информации о переводе монет
type TransferInfo struct {
	FromUser string `json:"fromUser"` // Имя пользователя, который отправил монеты
	ToUser   string `json:"toUser"`   // Имя пользователя, которому отправлены монеты
	Amount   int    `json:"amount"`   // Количество переведенных монет
}

// InfoResponse - структура для ответа на запрос информации о монетах и инвентаре
type InfoResponse struct {
	Coins      int            `json:"coins"`      // Количество монет у пользователя
	Inventory  []InventoryItem `json:"inventory"`  // Инвентарь пользователя
	CoinHistory CoinHistory   `json:"coinHistory"` // История монет
}

// Merchandise - структура для товара
type Merchandise struct {
	ID    int    `json:"id"`    // Уникальный идентификатор товара
	Name  string `json:"name"`  // Название товара
	Price int    `json:"price"` // Цена товара в монетах
}

// User - структура для пользователя
type User struct {
	ID                int    `json:"id"`                // Уникальный идентификатор пользователя
	Username          string `json:"username"`          // Имя пользователя
	PasswordHash      string `json:"-"`                 // Хэш пароля (не возвращаем в ответах)
	Coins             int    `json:"coins"`             // Количество монет у пользователя
	PurchasedMerch    []Merchandise // Список купленных товаров
	IncomingTransfers []TransferInfo // Список входящих переводов
	OutgoingTransfers []TransferInfo // Список исходящих переводов
}

// TransferCoins - метод для перевода монет от одного пользователя другому
func (u *User) TransferCoins(recipient *User, coins int) error {
    if u.Coins < coins {
        return fmt.Errorf("недостаточно монет для перевода")
    }
    u.Coins -= coins
    recipient.Coins += coins
    // Запись в историю переводов, если необходимо
    // Можно добавлять логику для записи транзакции в базу данных
    return nil
}

// BuyMerch - метод для покупки товара пользователем
func (u *User) BuyMerch(item *Merchandise) error {
    if u.Coins < item.Price {
        return fmt.Errorf("недостаточно монет для покупки")
    }
    u.Coins -= item.Price
    // Добавить товар в список покупок пользователя
    u.PurchasedMerch = append(u.PurchasedMerch, *item)
    // Логика для записи покупки в базу данных, если необходимо
    return nil
}


func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, coins FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Coins)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

func CreateUser(username, password string) (*User, error) {
	// Генерация хеша пароля должна быть сделана безопасно
	// Для примера, не будем добавлять хеширование
	_, err := db.Exec("INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3)", username, password, 1000)
	if err != nil {
		return nil, err
	}
	return GetUserByUsername(username)
}

func GetMerchandiseByName(name string) (*Merchandise, error) {
	var item Merchandise
	err := db.QueryRow("SELECT id, name, price FROM merchandise WHERE name = $1", name).Scan(&item.ID, &item.Name, &item.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("товар не найден")
		}
		return nil, err
	}
	return &item, nil
}



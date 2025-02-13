package main

import (
	"errors"
	"sync"
)

// Определение структуры товара (мерча)
type Merch struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// Каталог мерча (10 видов товаров)
var MerchCatalog = map[string]Merch{
	"t-shirt":    {"t-shirt", 80},
	"cup":        {"cup", 20},
	"book":       {"book", 50},
	"pen":        {"pen", 10},
	"powerbank":  {"powerbank", 200},
	"hoody":      {"hoody", 300},
	"umbrella":   {"umbrella", 200},
	"socks":      {"socks", 10},
	"wallet":     {"wallet", 50},
	"pink-hoody": {"pink-hoody", 500},
}

// User представляет сотрудника
type User struct {
	Username          string         `json:"username"`
	Coins             int            `json:"coins"`
	PurchasedMerch    []Merch        `json:"purchased_merch"`
	IncomingTransfers []TransferInfo `json:"incoming_transfers"`
	OutgoingTransfers []TransferInfo `json:"outgoing_transfers"`
	mu                sync.Mutex
}

// TransferInfo хранит информацию о переводе монет
type TransferInfo struct {
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
	Coins int    `json:"coins"`
}

var (
	users   = make(map[string]*User)
	usersMu sync.RWMutex
)

// Инициализировать хранилище
func InitStore() {
	// При старте можно инициализировать нескольких сотрудников
}

// GetOrCreateUser возвращает пользователя, создавая его, если не существует
func GetOrCreateUser(username string) *User {
	usersMu.Lock()
	defer usersMu.Unlock()
	if user, ok := users[username]; ok {
		return user
	}
	// Каждый новый сотрудник получает 1000 монет
	user := &User{
		Username:       username,
		Coins:          1000,
		PurchasedMerch: []Merch{},
	}
	users[username] = user
	return user
}

// GetUser возвращает пользователя, если существует
func GetUser(username string) *User {
	usersMu.RLock()
	defer usersMu.RUnlock()
	return users[username]
}

// BuyMerch покупает мерч, списывая монетки и добавляя товар в список покупок
func (u *User) BuyMerch(item Merch) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.Coins < item.Price {
		return errors.New("недостаточно монет")
	}
	u.Coins -= item.Price
	u.PurchasedMerch = append(u.PurchasedMerch, item)
	return nil
}

// TransferCoins переводит монетки от пользователя к получателю
func (u *User) TransferCoins(recipient *User, coins int) error {
	// Блокировка sender и recipient для безопасного обновления
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.Coins < coins {
		return errors.New("недостаточно монет для перевода")
	}
	u.Coins -= coins
	// Обновляем исходящие переводы у отправителя
	u.OutgoingTransfers = append(u.OutgoingTransfers, TransferInfo{To: recipient.Username, Coins: coins})

	// Обновляем получателя
	recipient.mu.Lock()
	defer recipient.mu.Unlock()
	recipient.Coins += coins
	recipient.IncomingTransfers = append(recipient.IncomingTransfers, TransferInfo{From: u.Username, Coins: coins})
	return nil
}


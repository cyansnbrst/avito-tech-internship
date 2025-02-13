package models

import "time"

// User model
type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Balance      int64     `db:"balance"`
	CreatedAt    time.Time `db:"created_at"`
}

// User info response model
type InfoResponse struct {
	CoinsInventory
	CoinHistory *TransactionHistory `json:"coin_history"`
}

// Coins and inventory
type CoinsInventory struct {
	Coins     int64           `json:"coins"`
	Inventory []InventoryItem `json:"inventory"`
}

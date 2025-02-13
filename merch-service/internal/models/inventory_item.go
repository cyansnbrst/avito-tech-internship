package models

// Inventory item struct
type InventoryItem struct {
	Type     string `db:"type" json:"type"`
	Quantity int64  `db:"quantity" json:"quantity"`
}

package models

// Transaction history struct
type TransactionHistory struct {
	Received []ReceiveTransaction `json:"received"`
	Sent     []SendTransaction    `json:"sent"`
}

// Receive transaction struct
type ReceiveTransaction struct {
	FromUser string `db:"from_user" json:"from_user"`
	Amount   int64  `db:"amount" json:"amount"`
}

// Send transaction struct
type SendTransaction struct {
	ToUser string `db:"to_user" json:"to_user" validate:"required"`
	Amount int64  `db:"amount" json:"amount" validate:"required,min=1"`
}

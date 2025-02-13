package models

import "time"

// Product struct
type Product struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Price     int64     `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

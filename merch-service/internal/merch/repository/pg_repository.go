package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cyansnbrst/merch-service/internal/merch"
	m "cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
)

// Merch repository struct
type merchRepo struct {
	db *pgxpool.Pool
}

// Merch repository constructor
func NewMerchRepo(db *pgxpool.Pool) merch.Repository {
	return &merchRepo{db: db}
}

// Get user's main info
func (r *merchRepo) GetCoinsAndInventory(ctx context.Context, userID int64) (*m.CoinsInventory, error) {
	query := `
		SELECT u.balance, it.name AS item_type, ii.quantity
		FROM users u
		LEFT JOIN inventory_items ii ON u.id = ii.user_id
		LEFT JOIN items it ON ii.item_id = it.id
		WHERE u.id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo - failed to execute query: %w", err)
	}
	defer rows.Close()

	var coins int64
	inventory := make([]m.InventoryItem, 0)
	for rows.Next() {
		var itemType *string
		var quantity *int64
		if err := rows.Scan(&coins, &itemType, &quantity); err != nil {
			return nil, fmt.Errorf("repo - failed to scan row: %w", err)
		}

		if itemType != nil {
			inventory = append(inventory, m.InventoryItem{
				Type:     *itemType,
				Quantity: *quantity,
			})
		}
	}

	return &m.CoinsInventory{
		Coins:     coins,
		Inventory: inventory,
	}, nil
}

// Get transaction history
func (r *merchRepo) GetTransactionHistory(ctx context.Context, userID int64) (*m.TransactionHistory, error) {
	query := `
		SELECT 'received' as type, u.username, t.amount 
		FROM transactions t 
		JOIN users u ON t.from_id = u.id 
		WHERE t.to_id = $1
		UNION ALL
		SELECT 'sent', u.username, t.amount 
		FROM transactions t 
		JOIN users u ON t.to_id = u.id 
		WHERE t.from_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repo - failed to get transactions: %w", err)
	}
	defer rows.Close()

	history := &m.TransactionHistory{
		Received: make([]m.ReceiveTransaction, 0),
		Sent:     make([]m.SendTransaction, 0),
	}

	for rows.Next() {
		var (
			txType   string
			username string
			amount   int64
		)

		if err := rows.Scan(&txType, &username, &amount); err != nil {
			return nil, fmt.Errorf("repo - failed to scan transaction: %w", err)
		}

		switch txType {
		case "received":
			history.Received = append(history.Received, m.ReceiveTransaction{
				FromUser: username,
				Amount:   amount,
			})
		case "sent":
			history.Sent = append(history.Sent, m.SendTransaction{
				ToUser: username,
				Amount: amount,
			})
		}
	}

	return history, nil
}

// Send coins to other user
func (r *merchRepo) SendCoins(ctx context.Context, fromUser int64, toUser string, amount int64) error {
	return r.execTx(ctx, func(tx pgx.Tx) error {
		balanceQuery := `
			SELECT u1.balance, u2.id
			FROM users u1
			JOIN users u2 ON u2.username = $1
			WHERE u1.id = $2
		`
		var currentBalance, toUserID int64
		err := tx.QueryRow(ctx, balanceQuery, toUser, fromUser).Scan(&currentBalance, &toUserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return db.ErrUserNotFound
			}
			return fmt.Errorf("repo - failed to get balance and recipient: %w", err)
		}

		if fromUser == toUserID {
			return db.ErrIncorrectReciever
		}

		if currentBalance < amount {
			return db.ErrInsufficientFunds
		}

		if err := r.updateBalance(ctx, tx, fromUser, -amount); err != nil {
			return err
		}

		if err := r.updateBalance(ctx, tx, toUserID, amount); err != nil {
			return err
		}

		if err := r.recordTransaction(ctx, tx, fromUser, toUserID, amount); err != nil {
			return err
		}

		return nil
	})
}

// Buy an item
func (r *merchRepo) BuyItem(ctx context.Context, userID int64, itemName string) error {
	return r.execTx(ctx, func(tx pgx.Tx) error {
		query := `
			SELECT i.id, i.price, u.balance
			FROM items i
			JOIN users u ON u.id = $1
			WHERE i.name = $2
		`
		var itemID, price, balance int64
		err := tx.QueryRow(ctx, query, userID, itemName).Scan(&itemID, &price, &balance)
		if err != nil {
			if err == pgx.ErrNoRows {
				return db.ErrItemtNotFound
			}
			return fmt.Errorf("repo - failed to get item and balance: %w", err)
		}

		if balance < price {
			return db.ErrInsufficientFunds
		}

		if err := r.updateBalance(ctx, tx, userID, -price); err != nil {
			return err
		}

		upsertInventoryQuery := `
			INSERT INTO inventory_items (user_id, item_id, quantity)
			VALUES ($1, $2, 1)
			ON CONFLICT (user_id, item_id) DO UPDATE SET quantity = inventory_items.quantity + 1
		`
		_, err = tx.Exec(ctx, upsertInventoryQuery, userID, itemID)
		if err != nil {
			return fmt.Errorf("repo - failed to update inventory: %w", err)
		}

		return nil
	})
}

// Execute a transaction
func (r *merchRepo) execTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo - failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if err != pgx.ErrTxClosed {
				log.Printf("repo - failed to rollback transaction: %v", err)
			}
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repo - failed to commit transaction: %w", err)
	}

	return nil
}

// Update user's balance
func (r *merchRepo) updateBalance(ctx context.Context, tx pgx.Tx, userID, amount int64) error {
	query := `
		UPDATE users
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err := tx.Exec(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("repo - failed to update balance: %w", err)
	}
	return nil
}

// Record coin transaction
func (r *merchRepo) recordTransaction(ctx context.Context, tx pgx.Tx, fromUser, toUser, amount int64) error {
	query := `
		INSERT INTO transactions (from_id, to_id, amount)
		VALUES ($1, $2, $3)
	`
	_, err := tx.Exec(ctx, query, fromUser, toUser, amount)
	if err != nil {
		return fmt.Errorf("repo - failed to record transaction: %w", err)
	}
	return nil
}

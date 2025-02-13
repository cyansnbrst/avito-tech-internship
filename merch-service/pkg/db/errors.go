package db

import "errors"

var (
	ErrItemtNotFound     = errors.New("item not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrIncorrectReciever = errors.New("can't send money to the same user")
	ErrUserNotFound      = errors.New("user not found")
)

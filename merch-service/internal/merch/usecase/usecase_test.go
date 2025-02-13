package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mock_merch "cyansnbrst/merch-service/internal/merch/mock"
	"cyansnbrst/merch-service/internal/merch/usecase"
	m "cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
)

var ErrRandomDBError = errors.New("db error")

func TestMerchUC_GetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo)

	tests := []struct {
		name          string
		userID        int64
		mockSetup     func()
		expectedResp  *m.InfoResponse
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			mockSetup: func() {
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(1)).Return(&m.CoinsInventory{
					Coins: 100,
					Inventory: []m.InventoryItem{
						{
							Type:     "hoodie",
							Quantity: 2,
						},
						{
							Type:     "wallet",
							Quantity: 1,
						},
					},
				}, nil)
				mockRepo.EXPECT().GetTransactionHistory(gomock.Any(), int64(1)).Return(&m.TransactionHistory{
					Received: []m.ReceiveTransaction{
						{FromUser: "user1", Amount: 50},
					},
					Sent: []m.SendTransaction{
						{ToUser: "user2", Amount: 100},
					},
				}, nil)
			},
			expectedResp: &m.InfoResponse{
				CoinsInventory: m.CoinsInventory{
					Coins: 100,
					Inventory: []m.InventoryItem{
						{
							Type:     "hoodie",
							Quantity: 2,
						},
						{
							Type:     "wallet",
							Quantity: 1,
						},
					},
				},
				CoinHistory: &m.TransactionHistory{
					Received: []m.ReceiveTransaction{
						{FromUser: "user1", Amount: 50},
					},
					Sent: []m.SendTransaction{
						{ToUser: "user2", Amount: 100},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:   "error in get coins and inventory",
			userID: 2,
			mockSetup: func() {
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(2)).Return(nil, ErrRandomDBError)
			},
			expectedResp:  nil,
			expectedError: ErrRandomDBError,
		},
		{
			name:   "error in get transaction history",
			userID: 3,
			mockSetup: func() {
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(3)).Return(&m.CoinsInventory{
					Coins:     200,
					Inventory: []m.InventoryItem{},
				}, nil)
				mockRepo.EXPECT().GetTransactionHistory(gomock.Any(), int64(3)).Return(nil, ErrRandomDBError)
			},
			expectedResp:  nil,
			expectedError: ErrRandomDBError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := merchUC.GetInfo(context.Background(), tt.userID)

			assert.Equal(t, tt.expectedResp, resp)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestSendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo)

	tests := []struct {
		name          string
		fromUserID    int64
		toUser        string
		amount        int64
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "success",
			fromUserID: 1,
			toUser:     "user1",
			amount:     50,
			mockSetup: func() {
				mockRepo.EXPECT().SendCoins(gomock.Any(), int64(1), "user1", int64(50)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "error insufficient funds",
			fromUserID: 1,
			toUser:     "user2",
			amount:     200,
			mockSetup: func() {
				mockRepo.EXPECT().SendCoins(gomock.Any(), int64(1), "user2", int64(200)).Return(db.ErrInsufficientFunds)
			},
			expectedError: db.ErrInsufficientFunds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := merchUC.SendCoins(context.Background(), tt.fromUserID, tt.toUser, tt.amount)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo)

	tests := []struct {
		name          string
		userID        int64
		item          string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			item:   "item1",
			mockSetup: func() {
				mockRepo.EXPECT().BuyItem(gomock.Any(), int64(1), "item1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "error item not found",
			userID: 1,
			item:   "item1",
			mockSetup: func() {
				mockRepo.EXPECT().BuyItem(gomock.Any(), int64(1), "item1").Return(db.ErrItemtNotFound)
			},
			expectedError: db.ErrItemtNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := merchUC.BuyItem(context.Background(), tt.userID, tt.item)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

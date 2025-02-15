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
	"cyansnbrst/merch-service/pkg/db/redis"
)

var ErrRandomDBError = errors.New("db error")

func TestMerchUC_GetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)
	mockRedisRepo := mock_merch.NewMockRedisRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo, mockRedisRepo)

	tests := []struct {
		name          string
		userID        int64
		mockSetup     func()
		expectedResp  *m.InfoResponse
		expectedError error
	}{
		{
			name:   "success data from DB",
			userID: 1,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, nil)
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
				mockRedisRepo.EXPECT().SetInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
			name:   "success data from redis",
			userID: 2,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(&m.InfoResponse{
					CoinsInventory: m.CoinsInventory{
						Coins: 200,
						Inventory: []m.InventoryItem{
							{
								Type:     "t-shirt",
								Quantity: 3,
							},
						},
					},
					CoinHistory: &m.TransactionHistory{
						Received: []m.ReceiveTransaction{
							{FromUser: "user3", Amount: 75},
						},
						Sent: []m.SendTransaction{
							{ToUser: "user4", Amount: 50},
						},
					},
				}, nil)
			},
			expectedResp: &m.InfoResponse{
				CoinsInventory: m.CoinsInventory{
					Coins: 200,
					Inventory: []m.InventoryItem{
						{
							Type:     "t-shirt",
							Quantity: 3,
						},
					},
				},
				CoinHistory: &m.TransactionHistory{
					Received: []m.ReceiveTransaction{
						{FromUser: "user3", Amount: 75},
					},
					Sent: []m.SendTransaction{
						{ToUser: "user4", Amount: 50},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:   "error redis error",
			userID: 3,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, ErrRandomDBError)
			},
			expectedResp:  nil,
			expectedError: ErrRandomDBError,
		},
		{
			name:   "error db error in GetCoinsAndInventory",
			userID: 4,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(4)).Return(nil, ErrRandomDBError)
			},
			expectedResp:  nil,
			expectedError: ErrRandomDBError,
		},
		{
			name:   "error db error in GetTransactionHistory",
			userID: 5,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(5)).Return(&m.CoinsInventory{
					Coins:     300,
					Inventory: []m.InventoryItem{},
				}, nil)
				mockRepo.EXPECT().GetTransactionHistory(gomock.Any(), int64(5)).Return(nil, ErrRandomDBError)
			},
			expectedResp:  nil,
			expectedError: ErrRandomDBError,
		},
		{
			name:   "error redis SetInfo error",
			userID: 6,
			mockSetup: func() {
				mockRedisRepo.EXPECT().GetInfo(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockRepo.EXPECT().GetCoinsAndInventory(gomock.Any(), int64(6)).Return(&m.CoinsInventory{
					Coins:     400,
					Inventory: []m.InventoryItem{},
				}, nil)
				mockRepo.EXPECT().GetTransactionHistory(gomock.Any(), int64(6)).Return(&m.TransactionHistory{
					Received: []m.ReceiveTransaction{},
					Sent:     []m.SendTransaction{},
				}, nil)
				mockRedisRepo.EXPECT().SetInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrRandomDBError)
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

func TestMerchUC_SendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)
	mockRedisRepo := mock_merch.NewMockRedisRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo, mockRedisRepo)

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
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(nil)
				mockRepo.EXPECT().GetUserIDByUsername(gomock.Any(), "user1").Return(int64(2), nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(2))).Return(nil)
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
		{
			name:       "error delete cache for sender",
			fromUserID: 1,
			toUser:     "user3",
			amount:     50,
			mockSetup: func() {
				mockRepo.EXPECT().SendCoins(gomock.Any(), int64(1), "user3", int64(50)).Return(nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
		},
		{
			name:       "error delete cache for receiver",
			fromUserID: 1,
			toUser:     "user4",
			amount:     50,
			mockSetup: func() {
				mockRepo.EXPECT().SendCoins(gomock.Any(), int64(1), "user4", int64(50)).Return(nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(nil)
				mockRepo.EXPECT().GetUserIDByUsername(gomock.Any(), "user4").Return(int64(3), nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(3))).Return(ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
		},
		{
			name:       "error get user ID by username",
			fromUserID: 1,
			toUser:     "user5",
			amount:     50,
			mockSetup: func() {
				mockRepo.EXPECT().SendCoins(gomock.Any(), int64(1), "user5", int64(50)).Return(nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(nil)
				mockRepo.EXPECT().GetUserIDByUsername(gomock.Any(), "user5").Return(int64(0), ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
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

func TestMerchUC_BuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_merch.NewMockRepository(ctrl)
	mockRedisRepo := mock_merch.NewMockRedisRepository(ctrl)

	merchUC := usecase.NewMerchUseCase(mockRepo, mockRedisRepo)

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
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(nil)
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
		{
			name:   "error delete cache",
			userID: 1,
			item:   "item1",
			mockSetup: func() {
				mockRepo.EXPECT().BuyItem(gomock.Any(), int64(1), "item1").Return(nil)
				mockRedisRepo.EXPECT().DeleteInfo(gomock.Any(), redis.GetUserInfoCacheKey(int64(1))).Return(ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
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

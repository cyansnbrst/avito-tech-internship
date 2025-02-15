package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"cyansnbrst/merch-service/internal/auth"
	"cyansnbrst/merch-service/internal/auth/repository"
	"cyansnbrst/merch-service/internal/auth/usecase"
	"cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/internal/server"
	"cyansnbrst/merch-service/pkg/db"
)

type MerchTestSuite struct {
	BaseTestSuite
	authUC auth.UseCase
}

func TestMerchSuite(t *testing.T) {
	suite.Run(t, new(MerchTestSuite))
}

func (s *MerchTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	authRepo := repository.NewAuthRepo(s.dbPool)
	s.authUC = usecase.NewAuthUseCase(s.cfg, authRepo)
}

func (s *MerchTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *MerchTestSuite) TestMerch_BuyItem_Success() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	user := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "asdlfkas2op2348n3",
	}
	item := "pink-hoody"
	itemID := 10

	var id, balance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, balance`,
		user.Username, user.PasswordHash,
	).Scan(&id, &balance)
	s.Require().NoError(err)
	s.Equal(1000, balance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(id)})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", ts.URL, item), nil)
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	err = s.dbPool.QueryRow(context.Background(),
		`SELECT balance FROM users 
		WHERE id = $1`,
		id,
	).Scan(&balance)
	s.Require().NoError(err)
	s.Equal(500, balance)

	var quantity int
	err = s.dbPool.QueryRow(context.Background(),
		`SELECT quantity FROM inventory_items 
		WHERE quantity = 1 AND item_id = $1`,
		itemID,
	).Scan(&quantity)
	s.Require().NoError(err)
	s.Equal(1, quantity)
}

func (s *MerchTestSuite) TestMerch_BuyItem_InsufficientFunds() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	user := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "asdlfkas2op2348n3",
	}
	item := "pink-hoody"

	var id, balance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, balance) 
		VALUES ($1, $2, $3) 
		RETURNING id, balance`,
		user.Username, user.PasswordHash, 300,
	).Scan(&id, &balance)
	s.Require().NoError(err)
	s.Equal(300, balance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(id)})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", ts.URL, item), nil)
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal(db.ErrInsufficientFunds.Error(), response["errors"])

	err = s.dbPool.QueryRow(context.Background(),
		`SELECT balance FROM users 
		WHERE id = $1`,
		id,
	).Scan(&balance)
	s.Require().NoError(err)
	s.Equal(300, balance)
}

func (s *MerchTestSuite) TestMerch_BuyItem_ItemNotFound() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	user := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "asdlfkas2op2348n3",
	}
	item := "non-existent-item"

	var id, balance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, balance`,
		user.Username, user.PasswordHash,
	).Scan(&id, &balance)
	s.Require().NoError(err)
	s.Equal(1000, balance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(id)})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", ts.URL, item), nil)
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal("item not found", response["errors"])

	err = s.dbPool.QueryRow(context.Background(),
		`SELECT balance FROM users 
		WHERE id = $1`,
		id,
	).Scan(&balance)
	s.Require().NoError(err)
	s.Equal(1000, balance)
}

func (s *MerchTestSuite) TestMerch_Unauthorized() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	item := "pink-hoody"

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", ts.URL, item), nil)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal("you must be authenticated to access this resource", response["errors"])
}

func (s *MerchTestSuite) TestMerch_SendCoins_Success() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	sender := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "sadfswergwrb",
	}
	var senderID, senderBalance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, balance`,
		sender.Username, sender.PasswordHash,
	).Scan(&senderID, &senderBalance)
	s.Require().NoError(err)
	s.Equal(1000, senderBalance)

	receiver := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "gasgtefgdagdsag",
	}
	var receiverID, receiverBalance int
	err = s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, balance) 
		VALUES ($1, $2, $3) 
		RETURNING id, balance`,
		receiver.Username, receiver.PasswordHash, 500,
	).Scan(&receiverID, &receiverBalance)
	s.Require().NoError(err)
	s.Equal(500, receiverBalance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(senderID)})
	s.Require().NoError(err)

	transferAmount := 300
	reqBody := fmt.Sprintf(`{"to_user": "%s", "amount": %d}`, receiver.Username, transferAmount)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	err = s.dbPool.QueryRow(context.Background(),
		`SELECT balance FROM users 
		WHERE id = $1`,
		senderID,
	).Scan(&senderBalance)
	s.Require().NoError(err)
	s.Equal(700, senderBalance)

	err = s.dbPool.QueryRow(context.Background(),
		`SELECT balance FROM users 
		WHERE id = $1`,
		receiverID,
	).Scan(&receiverBalance)
	s.Require().NoError(err)
	s.Equal(800, receiverBalance)

	var transactionID int
	var fromID, toID, amount int
	var transactionDate time.Time
	err = s.dbPool.QueryRow(context.Background(),
		`SELECT id, from_id, to_id, amount, transaction_date 
		FROM transactions 
		WHERE from_id = $1 AND to_id = $2 AND amount = $3`,
		senderID, receiverID, transferAmount,
	).Scan(&transactionID, &fromID, &toID, &amount, &transactionDate)
	s.Require().NoError(err)

	s.Equal(senderID, fromID)
	s.Equal(receiverID, toID)
	s.Equal(transferAmount, amount)
	s.WithinDuration(time.Now(), transactionDate, time.Second)
}

func (s *MerchTestSuite) TestMerch_SendCoins_UserNotFound() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	sender := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "sadfswergwrb",
	}
	var senderID, senderBalance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id, balance`,
		sender.Username, sender.PasswordHash,
	).Scan(&senderID, &senderBalance)
	s.Require().NoError(err)
	s.Equal(1000, senderBalance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(senderID)})
	s.Require().NoError(err)

	transferAmount := 300
	reqBody := fmt.Sprintf(`{"to_user": "%s", "amount": %d}`, "some-user", transferAmount)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal(db.ErrUserNotFound.Error(), response["errors"])
}

func (s *MerchTestSuite) TestMerch_SendCoins_InsufficientFunds() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	sender := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "sadfswergwrb",
	}
	var senderID, senderBalance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, balance) 
        VALUES ($1, $2, $3) 
        RETURNING id, balance`,
		sender.Username, sender.PasswordHash, 200,
	).Scan(&senderID, &senderBalance)
	s.Require().NoError(err)
	s.Equal(200, senderBalance)

	receiver := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "gasgtefgdagdsag",
	}
	var receiverID, receiverBalance int
	err = s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, balance) 
        VALUES ($1, $2, $3) 
        RETURNING id, balance`,
		receiver.Username, receiver.PasswordHash, 500,
	).Scan(&receiverID, &receiverBalance)
	s.Require().NoError(err)
	s.Equal(500, receiverBalance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(senderID)})
	s.Require().NoError(err)

	transferAmount := 300
	reqBody := fmt.Sprintf(`{"to_user": "%s", "amount": %d}`, receiver.Username, transferAmount)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal(db.ErrInsufficientFunds.Error(), response["errors"])
}

func (s *MerchTestSuite) TestMerch_SendCoins_InvalidJSON() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	token, err := s.authUC.GenerateJWT(&models.User{ID: 1})
	s.Require().NoError(err)

	reqBody := `{"to_user": user-123, "amount": 300}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *MerchTestSuite) TestMerch_SendCoins_NegativeAmount() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	sender := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "sadfswergwrb",
	}
	var senderID, senderBalance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id, balance`,
		sender.Username, sender.PasswordHash,
	).Scan(&senderID, &senderBalance)
	s.Require().NoError(err)
	s.Equal(1000, senderBalance)

	receiver := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "gasgtefgdagdsag",
	}
	var receiverID, receiverBalance int
	err = s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id, balance`,
		receiver.Username, receiver.PasswordHash,
	).Scan(&receiverID, &receiverBalance)
	s.Require().NoError(err)
	s.Equal(1000, receiverBalance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(senderID)})
	s.Require().NoError(err)

	transferAmount := -300
	reqBody := fmt.Sprintf(`{"to_user": "%s", "amount": %d}`, receiver.Username, transferAmount)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal("field 'Amount' failed on the 'min' rule", response["errors"])
}

func (s *MerchTestSuite) TestMerch_SendCoins_SelfTransfer() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	user := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "sadfswergwrb",
	}
	var userID, userBalance int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id, balance`,
		user.Username, user.PasswordHash,
	).Scan(&userID, &userBalance)
	s.Require().NoError(err)
	s.Equal(1000, userBalance)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(userID)})
	s.Require().NoError(err)

	transferAmount := 300
	reqBody := fmt.Sprintf(`{"to_user": "%s", "amount": %d}`, user.Username, transferAmount)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/sendCoin", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal(db.ErrIncorrectReciever.Error(), response["errors"])
}

func (s *MerchTestSuite) TestMerch_GetInfo_Success() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool, s.redisClient)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	user := models.User{
		Username:     "user-" + uuid.New().String(),
		PasswordHash: "password-hash",
	}
	var userID int
	err := s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id`,
		user.Username, user.PasswordHash,
	).Scan(&userID)
	s.Require().NoError(err)

	receiver := models.User{
		Username:     "receiver-" + uuid.New().String(),
		PasswordHash: "receiver-password-hash",
	}
	var receiverID int
	err = s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id`,
		receiver.Username, receiver.PasswordHash,
	).Scan(&receiverID)
	s.Require().NoError(err)

	sender := models.User{
		Username:     "sender-" + uuid.New().String(),
		PasswordHash: "sender-password-hash",
	}
	var senderID int
	err = s.dbPool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id`,
		sender.Username, sender.PasswordHash,
	).Scan(&senderID)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		`INSERT INTO inventory_items (user_id, item_id, quantity) 
        VALUES ($1, $2, $3)`,
		userID, 1, 1,
	)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		`INSERT INTO transactions (from_id, to_id, amount) 
        VALUES ($1, $2, $3), ($4, $5, $6)`,
		userID, receiverID, 200,
		senderID, userID, 300,
	)
	s.Require().NoError(err)

	token, err := s.authUC.GenerateJWT(&models.User{ID: int64(userID)})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/info", nil)
	s.Require().NoError(err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var infoResponse models.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	s.Require().NoError(err)

	s.Equal(int64(1000), infoResponse.Coins)

	s.Require().NotNil(infoResponse.Inventory)
	s.Require().Len(infoResponse.Inventory, 1)
	s.Equal(int64(1), infoResponse.Inventory[0].Quantity)

	s.Require().NotNil(infoResponse.CoinHistory)
	s.Require().Len(infoResponse.CoinHistory.Received, 1)
	s.Require().Len(infoResponse.CoinHistory.Sent, 1)

	receivedTx := infoResponse.CoinHistory.Received[0]
	s.Equal(sender.Username, receivedTx.FromUser)
	s.Equal(int64(300), receivedTx.Amount)

	sentTx := infoResponse.CoinHistory.Sent[0]
	s.Equal(receiver.Username, sentTx.ToUser)
	s.Equal(int64(200), sentTx.Amount)
}

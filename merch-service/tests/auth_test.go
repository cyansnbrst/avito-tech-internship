package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexedwards/argon2id"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/internal/server"
)

type AuthTestSuite struct {
	BaseTestSuite
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
}

func (s *AuthTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *AuthTestSuite) TestAuth_Authenticate_Login() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	username := "user-" + uuid.New().String()[:8]
	password := "password"

	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		`INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)`,
		username, hashedPassword,
	)
	s.Require().NoError(err)

	reqBody := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/auth", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var authResp models.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	s.Require().NoError(err)
	s.NotEmpty(authResp.Token)
}

func (s *AuthTestSuite) TestAuth_Authenticate_InvalidPassword() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	username := "user-" + uuid.New().String()[:8]
	password := "password"

	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		`INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)`,
		username, hashedPassword,
	)
	s.Require().NoError(err)

	reqBody := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, "wrong-password")
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/auth", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	s.Require().NoError(err)
	s.Equal("invalid authentication credentials", response["errors"])
}

func (s *AuthTestSuite) TestAuth_Authenticate_Register() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	username := "user-" + uuid.New().String()[:8]
	password := "password"

	reqBody := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/auth", strings.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var authResp models.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	s.Require().NoError(err)
	s.NotEmpty(authResp.Token)

	var dbUser models.User
	err = s.dbPool.QueryRow(context.Background(),
		`SELECT username, password_hash FROM users 
		WHERE username = $1`,
		username,
	).Scan(&dbUser.Username, &dbUser.PasswordHash)
	s.Require().NoError(err)
	s.Equal(username, dbUser.Username)

	match, err := argon2id.ComparePasswordAndHash(password, dbUser.PasswordHash)
	s.Require().NoError(err)
	s.True(match)
}

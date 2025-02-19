package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	// swagger docs
	_ "cyansnbrst/merch-service/docs"
	authHTTP "cyansnbrst/merch-service/internal/auth/delivery/http"
	authRepository "cyansnbrst/merch-service/internal/auth/repository"
	authUseCase "cyansnbrst/merch-service/internal/auth/usecase"
	merchHTTP "cyansnbrst/merch-service/internal/merch/delivery/http"
	merchRepository "cyansnbrst/merch-service/internal/merch/repository"
	merchUseCase "cyansnbrst/merch-service/internal/merch/usecase"
	mm "cyansnbrst/merch-service/internal/middleware"
	"cyansnbrst/merch-service/pkg/validator"
)

// Register server handlers
func (s *Server) RegisterHandlers() *echo.Echo {
	e := echo.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Use(middleware.Recover())

	e.Validator = validator.NewCustomValidator()

	authRepo := authRepository.NewAuthRepo(s.db)
	merchRepo := merchRepository.NewMerchRepo(s.db)
	merchRedisRepo := merchRepository.NewMerchRedisRepo(s.config, s.redisClient)

	authUC := authUseCase.NewAuthUseCase(s.config, authRepo)
	merchUC := merchUseCase.NewMerchUseCase(merchRepo, merchRedisRepo)

	authHandlers := authHTTP.NewAuthHandlers(authUC, s.logger)
	merchHandlers := merchHTTP.NewMerchHandlers(merchUC, s.logger)

	mw := mm.NewManager(s.config, s.logger)

	api := e.Group("/api")
	protectedAPI := api.Group("")

	protectedAPI.Use(mw.Authenticate)

	authHTTP.RegisterAuthRoutes(api, authHandlers)
	merchHTTP.RegisterMerchRoutes(protectedAPI, merchHandlers)

	return e
}

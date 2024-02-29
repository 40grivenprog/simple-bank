package api

import (
	"fmt"

	db "github.com/40grivenprog/simple-bank/db/sqlc"
	"github.com/40grivenprog/simple-bank/token"
	"github.com/40grivenprog/simple-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	if gin.Mode() != gin.TestMode {
		router.Use(tracingMiddleware())
	}
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.POST("/credit_requests", server.createCreditRequest)
	authRoutes.GET("/credit_requests", server.listCreditRequests)

	adminRoutes := router.Group("/admin/").Use(authMiddleware(server.tokenMaker), roleMiddleware())
	adminRoutes.GET("/accounts", server.listAccounts)
	adminRoutes.GET("/credit_requests", server.listPengingCreditRequest)
	adminRoutes.PATCH("/credit_requests/:id", server.cancelPendingRequest)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

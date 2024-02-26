package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/40grivenprog/simple-bank/db/sqlc"
	"github.com/40grivenprog/simple-bank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type CreateCreditRequest struct {
	Reason   string `json:"reason"`
	Amount   int32  `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createCreditRequest(ctx *gin.Context) {
	var req CreateCreditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	err := server.validCreditRequest(ctx, authPayload.Username, req.Currency)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateCreditRequestParams{
		Username: authPayload.Username,
		Reason: sql.NullString{
			String: req.Reason,
		},
		Amount:   req.Amount,
		Currency: req.Currency,
	}

	createdCreditRequest, err := server.store.CreateCreditRequest(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(errors.New("user already have pending request")))
			}
		} else {
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, createdCreditRequest)
}

func (server *Server) listCreditRequests(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	creditRequests, err := server.store.GetCreditRequestsByUsername(ctx, authPayload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, creditRequests)
}

func (server *Server) listPengingCreditRequest(ctx *gin.Context) {
	creditRequests, err := server.store.GetUsersPendingCreditRequests(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, creditRequests)
}

type CancelCreditRequest struct {
	ID int64 `uri:"id" binding:"required,gte=0"`
}

func (server *Server) cancelPendingRequest(ctx *gin.Context) {
	var req CancelCreditRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.validCreditRequestForCancel(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	cancelledCreditRequest, err := server.store.CancelCreditRequestById(ctx, req.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cancelledCreditRequest)
}

func (server *Server) validCreditRequest(ctx *gin.Context, username, currency string) error {
	_, err := server.store.GetAccountByUsernameAndCurrency(ctx, db.GetAccountByUsernameAndCurrencyParams{
		Owner:    username,
		Currency: currency,
	})
	if err != nil {
		return fmt.Errorf("user haven't got account with currency: %s", currency)
	}
	return nil
}

func (server *Server) validCreditRequestForCancel(ctx *gin.Context, id int64) error {
	_, err := server.store.GetPendingCreditRequestById(ctx, id)
	if err != nil {
		return fmt.Errorf("no pending requests with id: %d", id)
	}
	return nil
}

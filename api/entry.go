package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/40grivenprog/simple-bank/db/sqlc"
	"github.com/40grivenprog/simple-bank/token"
	"github.com/gin-gonic/gin"
)

type listEntriesURIParams struct {
	AccountID int64 `uri:"account_id" binding:"required"`
}

type listEntriesQueryParams struct {
	PageID   int32 `form:"page_id"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var uriParams listEntriesURIParams
	fmt.Print(ctx.Params)
	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var queryParams listEntriesQueryParams
	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, uriParams.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account does not belong to auth user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListEntriesParams{
		Limit:     queryParams.PageSize,
		Offset:    queryParams.PageID,
		AccountID: uriParams.AccountID,
	}
	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, entries)
}

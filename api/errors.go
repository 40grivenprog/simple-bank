package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func handleEror(ctx *gin.Context, err error) {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code.Name() {
		case "foreign_key_violation", "unique_violation":
			ctx.JSON(http.StatusForbidden, errorResponse(errors.New("Account Already Exsists")))
		}
	} else if err == sql.ErrNoRows {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New("Invalid account")))
	} else {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

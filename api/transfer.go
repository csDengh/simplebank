package api

import (
	"database/sql"
	"net/http"

	"fmt"

	"errors"

	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/token"
	"github.com/gin-gonic/gin"
)

type CreateTranferReq struct {
	FromAccountID int64  `json:"from_account_id" binding:"required"`
	ToAccountID   int64  `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) CreateTransfer(ctx *gin.Context) {
	var req CreateTranferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	fromAccunt, ok := s.validAccount(ctx, req.FromAccountID, req.Currency)
	if !ok {
		return
	}

	if _, ok := s.validAccount(ctx, req.ToAccountID, req.Currency); !ok {
		return
	}

	authPlayLoader := ctx.MustGet(authorizationPayloadKey).(*token.PlayLoad)
	if authPlayLoader.Username != fromAccunt.Owner {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	args := db.TransferTxParams{
		FromAccountId: req.FromAccountID,
		ToAccountId:   req.ToAccountID,
		Amount:        req.Amount,
	}

	transfer := s.store.TransferTx(ctx, &args)
	if transfer.Err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(transfer.Err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

func (s *Server) validAccount(ctx *gin.Context, aId int64, currency string) (db.Account, bool) {

	a, err := s.store.GetAccount(ctx, aId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, ErrorRes(err))
			return a, false
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return a, false
	}

	if a.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", a.ID, a.Currency, currency)
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return a, false
	}

	return a, true
}

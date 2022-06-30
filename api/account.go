package api

import (
	"database/sql"
	"net/http"

	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type CreateAccountReq struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (s *Server) CreateAccount(ctx *gin.Context) {
	var req CreateAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	authPlayLoader := ctx.MustGet(authorizationPayloadKey).(*token.PlayLoad)

	args := db.CreateAccountParams{
		Owner:    authPlayLoader.Username,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := s.store.CreateAccount(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, ErrorRes(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	ctx.JSON(http.StatusOK, account)

}

type GetAccountReq struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) GetAccount(ctx *gin.Context) {
	var req GetAccountReq

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}

	account, err := s.store.GetAccount(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, ErrorRes(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}
	authPlayLoader := ctx.MustGet(authorizationPayloadKey).(*token.PlayLoad)
	if authPlayLoader.Username != account.Owner {
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type GetAccountListReq struct {
	PageId   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (s *Server) GetAccountList(ctx *gin.Context) {
	var req GetAccountListReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorRes(err))
		return
	}
	authPlayLoader := ctx.MustGet(authorizationPayloadKey).(*token.PlayLoad)

	accountList, err := s.store.ListAccounts(ctx, db.ListAccountsParams{
		Owner: authPlayLoader.Username,
		Limit:  req.PageSize,
		Offset: (req.PageId - 1) * req.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	ctx.JSON(http.StatusOK, accountList)
}

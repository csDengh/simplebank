package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) renewAccessToken(ctx *gin.Context) {

	var req renewAccessTokenReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	refpl, err := s.tokenMaker.ValidToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err)
		return
	}

	session, err := s.store.GetSession(ctx, refpl.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, ErrorRes(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	if session.Username != refpl.Username {
		err := fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	acctoken, accpl, err := s.tokenMaker.CreateToken(refpl.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	res := renewAccessTokenResponse{
		AccessToken:          acctoken,
		AccessTokenExpiresAt: accpl.TimeExprieAt,
	}
	ctx.JSON(http.StatusOK, res)
}

package api

import (
	"database/sql"
	"net/http"

	"time"

	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)


type CreateUserReq struct {
	Username  string `json:"username" binding:"required,alphanum"`
	Password  string `json:"password" binding:"required,min=6"`
	Full_name string `json:"full_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

func (s *Server) CreateUser(ctx *gin.Context) {
	var req CreateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	hashPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	args := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPwd,
		FullName:       req.Full_name,
		Email:          req.Email,
	}

	user, err := s.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, ErrorRes(err))
				return
			}
		}
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	user.HashedPassword = ""
	ctx.JSON(http.StatusOK, user)
}

type UserLoginReq struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLoginRes struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`

	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
}

func (s *Server) UserLogin(ctx *gin.Context) {

	var req UserLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	u, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, ErrorRes(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	err = utils.ConfirmPwd(u.HashedPassword, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorRes(err))
		return
	}

	accessToken, accessPlayload, err := s.tokenMaker.CreateToken(req.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	refleshToken, refleshPlayload , err := s.tokenMaker.CreateToken(req.Username, s.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	_, err = s.store.CreateSession(ctx, db.CreateSessionParams{
		ID: refleshPlayload.Id,
		Username: req.Username,
		RefreshToken: refleshToken,
		UserAgent: ctx.Request.UserAgent(),
		ClientIp: ctx.ClientIP(),
		IsBlocked: false,
		ExpiresAt: refleshPlayload.TimeExprieAt,
	})
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, ErrorRes(err))
		return
	}

	userRes := UserLoginRes{
		Username:          u.Username,
		FullName:          u.FullName,
		Email:             u.Email,
		PasswordChangedAt: u.PasswordChangedAt,
		CreatedAt:         u.CreatedAt,
		AccessToken:         accessToken,
		AccessTokenExpiresAt: accessPlayload.TimeExprieAt,

		SessionID: refleshPlayload.Id,
		RefreshToken: refleshToken,
		RefreshTokenExpiresAt: refleshPlayload.TimeExprieAt,
	}
	ctx.JSON(http.StatusOK, userRes)
}

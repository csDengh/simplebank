package gapi

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/pb"
	"github.com/csdengh/cur_blank/utils"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {

	hashPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return &pb.CreateUserRes{}, fmt.Errorf("hashPassword error")
	}

	args := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPwd,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := s.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return &pb.CreateUserRes{}, fmt.Errorf("unique_violation")
			}
		}
		return &pb.CreateUserRes{}, fmt.Errorf("db createUser error")
	}
	user.HashedPassword = ""
	return &pb.CreateUserRes{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		},
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.UserLoginReq) (*pb.UserLoginRes, error) {

	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.UserLoginRes{}, fmt.Errorf("not user")
		}
		return &pb.UserLoginRes{}, fmt.Errorf("access db error")
	}

	err = utils.ConfirmPwd(user.HashedPassword, req.Password)
	if err != nil {
		return &pb.UserLoginRes{}, fmt.Errorf("confirm password errors")
	}

	accessToken, accessPlayload, err := s.tokenMaker.CreateToken(req.Username, s.config.AccessTokenDuration)
	if err != nil {
		return &pb.UserLoginRes{}, fmt.Errorf("create accessToken error")
	}

	refleshToken, refleshPlayload, err := s.tokenMaker.CreateToken(req.Username, s.config.RefreshTokenDuration)
	if err != nil {
		return &pb.UserLoginRes{}, fmt.Errorf("create refleshToken error")
	}

	exinfo := s.extractMetadata(ctx)

	_, err = s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refleshPlayload.Id,
		Username:     req.Username,
		RefreshToken: refleshToken,
		UserAgent:    exinfo.UserAgent,
		ClientIp:     exinfo.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refleshPlayload.TimeExprieAt,
	})
	if err != nil {
		return &pb.UserLoginRes{}, fmt.Errorf("create session error")
	}

	userRes := &pb.UserLoginRes{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		},
		SessionId:             refleshPlayload.Id.String(),
		AccessToken:           accessToken,
		RefreshToken:          refleshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPlayload.TimeExprieAt),
		RefreshTokenExpiresAt: timestamppb.New(refleshPlayload.TimeExprieAt),
	}
	return userRes, nil
}

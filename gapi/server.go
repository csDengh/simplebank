package gapi

import (
	"github.com/csdengh/cur_blank/pb"
	"github.com/csdengh/cur_blank/token"
	"github.com/csdengh/cur_blank/utils"

	db "github.com/csdengh/cur_blank/db/sqlc"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	tokenMaker token.Maker
	config     *utils.Config
}

func NewServer(config *utils.Config, store_ db.Store) (*Server, error) {
	pm, err := token.NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		return nil, err
	}

	s := &Server{
		store:      store_,
		tokenMaker: pm,
		config:     config,
	}

	return s, nil
}

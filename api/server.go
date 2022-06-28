package api

import (
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store db.Store
	route *gin.Engine
}

func NewServer(store_ db.Store) *Server {
	s := &Server{
		store: store_,
	}

	route := gin.Default()

	route.POST("/accounts", s.CreateAccount)
	route.GET("/accounts/:id", s.GetAccount)
	route.GET("/accounts", s.GetAccountList)

	s.route = route
	return s
}

func (s *Server) Start(addr string) error {
	return s.route.Run(addr)
}

func errorRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}

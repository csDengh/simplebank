package api

import (
	"github.com/csdengh/cur_blank/token"
	"github.com/csdengh/cur_blank/utils"

	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store      db.Store
	route      *gin.Engine
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", currencyValid)
	}

	s.SartRoute()
	return s, nil
}

func (s *Server) Start(addr string) error {
	return s.route.Run(addr)
}

func (s *Server) SartRoute() {
	route := gin.Default()

	rg := route.Group("/", AuthenticateMideware(s.tokenMaker))

	rg.POST("/accounts", s.CreateAccount)
	rg.GET("/accounts/:id", s.GetAccount)
	rg.GET("/accounts", s.GetAccountList)

	rg.POST("/transfers", s.CreateTransfer)

	route.POST("/users", s.CreateUser)
	route.POST("/users/login", s.UserLogin)

	route.POST("/tokens/renew_access", s.renewAccessToken)

	s.route = route
}

func ErrorRes(err error) gin.H {
	return gin.H{"error": err.Error()}
}

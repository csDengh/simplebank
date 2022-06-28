package main

import (
	"database/sql"
	"log"

	"github.com/csdengh/cur_blank/api"
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/utils"
	_ "github.com/lib/pq"
)

func main() {

	config, err := utils.GetConfig(".")
	if err != nil {
		log.Fatalln("config load error", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalln(err)
	}
	store := db.NewStore(conn)

	s := api.NewServer(store)
	err = s.Start(config.ADDR)
	if err != nil {
		log.Fatalln(err)
	}
}

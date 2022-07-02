package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/csdengh/cur_blank/api"
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/gapi"
	"github.com/csdengh/cur_blank/pb"
	"github.com/csdengh/cur_blank/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	go grpcHttpServer(config, store)
	grpcServer(config, store)

}

func restServer(config *utils.Config, store db.Store) {
	s, err := api.NewServer(config, store)
	if err != nil {
		log.Fatalln(err)
	}

	err = s.Start(config.ADDR)
	if err != nil {
		log.Fatalln(err)
	}
}

func grpcServer(config *utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	gs := grpc.NewServer()

	pb.RegisterSimpleBankServer(gs, server)
	reflection.Register(gs)

	listener, err := net.Listen("tcp", config.GRPC_ADDR)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = gs.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server:", err)
	}
}

func grpcHttpServer(config *utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(http.Dir("./doc/swagger/")))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.ADDR)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

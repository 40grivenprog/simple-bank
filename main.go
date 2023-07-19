package main

import (
	"database/sql"
	"log"

	"github.com/40grivenprog/simple-bank/api"
	db "github.com/40grivenprog/simple-bank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:49152/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn , err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("can not connect to db")
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

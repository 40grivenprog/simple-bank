package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/40grivenprog/simple-bank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB


func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("can not read config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("can not connect to db")
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}

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
var testDBInstance *testDB

type testDB struct {
	*sql.DB
}


func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("can not read config:", err)
	}

	testDBInstance, err = newTestDB(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("can not connect to db")
	}

	testQueries = New(testDBInstance)

	os.Exit(m.Run())
}

func newTestStore(db *testDB) Store {
	return &SQLStore{
		db:      db.DB,
		Queries: New(db),
	}
}

func newTestDB(dbDriver, dbSource string) (*testDB, error) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		return nil, err
	}

	return &testDB{conn}, nil
}
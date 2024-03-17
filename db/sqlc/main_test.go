package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

const (
	driverName     = "postgres"
	dataSourceName = "postgresql://root:secret@localhost:5432/simplebank?sslmode=disable"
)

func TestMain(m *testing.M) {

	var err error
	testDB, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}

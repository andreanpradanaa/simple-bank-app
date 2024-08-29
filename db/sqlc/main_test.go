package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/jackc/pgx/v5/pgxpool"

)

var testStore Store

func TestMain(m *testing.M) {

	cfg, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database", err)
	}

	testStore = NewStore(conPool)

	os.Exit(m.Run())
}

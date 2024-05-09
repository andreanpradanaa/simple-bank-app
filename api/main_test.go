package api

import (
	"testing"
	"time"

	db "github.com/andreanpradanaa/simple-bank-app/db/sqlc"
	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

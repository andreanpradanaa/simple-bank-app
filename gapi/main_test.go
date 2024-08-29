package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/andreanpradanaa/simple-bank-app/db/sqlc"
	"github.com/andreanpradanaa/simple-bank-app/token"
	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/andreanpradanaa/simple-bank-app/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, worker worker.TaskDistributor) *Server {
	config := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, worker)
	require.NoError(t, err)

	return server
}

func NewContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, role string, duration time.Duration) context.Context {
	accesToken, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accesToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}

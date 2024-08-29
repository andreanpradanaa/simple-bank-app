package gapi

import (
	"fmt"

	db "github.com/andreanpradanaa/simple-bank-app/db/sqlc"
	"github.com/andreanpradanaa/simple-bank-app/pb"
	"github.com/andreanpradanaa/simple-bank-app/token"
	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/andreanpradanaa/simple-bank-app/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          utils.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config utils.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}

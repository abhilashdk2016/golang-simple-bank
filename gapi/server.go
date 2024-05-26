package gapi

import (
	"fmt"

	db "github.com/abhilashdk2016/golang-simple-bank/db/sqlc"
	"github.com/abhilashdk2016/golang-simple-bank/pb"
	"github.com/abhilashdk2016/golang-simple-bank/token"
	"github.com/abhilashdk2016/golang-simple-bank/util"
	"github.com/abhilashdk2016/golang-simple-bank/worker"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	pb.UnimplementedSimpleBankServer
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
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

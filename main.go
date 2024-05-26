package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	db "github.com/abhilashdk2016/golang-simple-bank/db/sqlc"
	"github.com/abhilashdk2016/golang-simple-bank/gapi"
	"github.com/abhilashdk2016/golang-simple-bank/mail"
	"github.com/abhilashdk2016/golang-simple-bank/pb"
	"github.com/abhilashdk2016/golang-simple-bank/util"
	"github.com/abhilashdk2016/golang-simple-bank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Unable to load Config")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database")
	}

	runDBMigration(config.MigrationUrl, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisServer,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	go runTaskProcessor(config, redisOpt, store)
	//runGinServer(store, config)
	go runGatewayServer(store, config, taskDistributor)
	runGRPCServer(store, config, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance: ", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up: ", err)
	}

	log.Println("DB Migrated Sucessfully")
}

// func runGinServer(store db.Store, config util.Config) {
// 	server, err := api.NewServer(config, store)
// 	if err != nil {
// 		log.Fatal("cannot create server", err)
// 	}
// 	err = server.Start(config.ServerAddress)
// 	if err != nil {
// 		log.Fatal("cannot start the server", err)
// 	}
// }

func runGRPCServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listner, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal("cannot create server", err)
	}
	fmt.Println("starting gRPC server at ", listner.Addr().String())
	err = grpcServer.Serve(listner)
	if err != nil {
		log.Fatal("cannot start grpc server", err)
	}
}

func runGatewayServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listner, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal("cannot create server", err)
	}
	fmt.Println("starting HTTP gRPC server at ", listner.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listner, handler)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server", err)
	}
}

func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Println("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal("failed to start task processor")
	}
}

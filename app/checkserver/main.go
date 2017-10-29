package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/api"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/store"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared/configuration"
	"github.com/dafanasiev/OTPCredentialProvider-backend/shared"
	"os/signal"
	"syscall"
)

func main() {
	selfDir, err := os.Getwd()
	if err!=nil {
		log.Fatal("cant get current working directory")
	}
	pathResolver := shared.NewPathResolver(selfDir, path.Join(selfDir, "../data"), path.Join(selfDir, "../etc"))
	configFileName := pathResolver.PathToAbs("${dir.config}/root.config")

	config, err := configuration.NewAppConfig(configFileName)
	if err != nil {
		log.Fatalf("fail to load config file [%v]; error:%v", configFileName, err.Error())
	}

	dbType := config.GetOrDie("db.type")

	dbConnectionString := config.GetOrDie("db.connectionString")

	db, err := store.NewUsersDb(dbType.(string), dbConnectionString.(string), pathResolver)
	if err != nil {
		log.Fatalf("db create failed: %v", err.Error())
	}

	apikey := config.GetOrDie("server.apikey")
	port := config.GetOrDie("server.port")
	host := config.GetOrDie("server.host")

	err  = db.Open()
	if err!=nil {
		log.Fatalf("unable to open db: %s", err.Error())
	}
	defer db.Close()

	hupC := make(chan os.Signal, 1)
	signal.Notify(hupC, syscall.SIGHUP)
	go func(){
		<-hupC
		db.Flush()
	}()

	bindTo := fmt.Sprintf("%s:%d", host, port)

	lis, err := net.Listen("tcp", bindTo)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		log.Printf("listen %s", bindTo)
	}

	s := grpc.NewServer()
	api.RegisterOTPCheckServer(s, &server{
		db:     db,
		apikey: apikey.(string),
	})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)
	go func(){
		<-ctrlC
		s.Stop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

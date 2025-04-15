package main

import (
	"context"
	"log"
	"net"

	"github.com/SoumyadipPayra/NightsWatch/src/db/conn"
	"github.com/SoumyadipPayra/NightsWatch/src/db/model"
	"github.com/SoumyadipPayra/NightsWatch/src/db/query"
	"github.com/SoumyadipPayra/NightsWatch/src/jwts"

	nwPB "github.com/SoumyadipPayra/NightsWatchProtobufs/gogenproto/nightswatch"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	nwPB.UnimplementedNightsWatchServiceServer
	queryEngine query.Query
}

func main() {
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	ctx = context.WithValue(ctx, "logger", logger)

	err := jwts.Initialize()
	if err != nil {
		log.Fatalf("failed to initialize jwts: %v", err)
	}

	err = conn.Initialize(ctx, &model.User{}, &model.DeviceData{})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	queryEngine := query.NewQuery(ctx)
	server := Server{
		queryEngine: queryEngine,
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	nwPB.RegisterNightsWatchServiceServer(s, &server)

	log.Printf("GRPC Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

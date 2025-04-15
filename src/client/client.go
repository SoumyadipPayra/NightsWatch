package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	nwPB "github.com/SoumyadipPayra/NightsWatchProtobufs/gogenproto/nightswatch"
)

func main() {
	ctx := context.Background()
	clientConn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer clientConn.Close()

	client := nwPB.NewNightsWatchServiceClient(clientConn)

	registerRequest := &nwPB.RegisterRequest{
		Name:     "testuser",
		Email:    "testuser@example.com",
		Password: "testpassword",
	}

	_, err = client.Register(ctx, registerRequest)
	if err != nil {
		log.Fatalf("Register failed: %v", err)
	}

	loginRequest := &nwPB.LoginRequest{
		Name:     "testuser",
		Password: "testpassword",
	}

	_, err = client.Login(ctx, loginRequest)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("Login successful")
}

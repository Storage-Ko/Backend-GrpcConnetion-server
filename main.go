package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Backend-Grpc-server/data"
	"google.golang.org/grpc"

	postpb "github.com/Backend-Grpc-server/protos/v1/post"
	userpb "github.com/Backend-Grpc-server/protos/v1/user"
)

const portNumber = "9001"

type postServer struct {
	postpb.PostServer

	userCli userpb.UserClient
}

var (
	once sync.Once
	cli  userpb.UserClient
)

func GetUserClient(serviceHost string) userpb.UserClient {
	once.Do(func() {
		conn, _ := grpc.Dial(serviceHost,
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)

		cli = userpb.NewUserClient(conn)
	})

	return cli
}

// ListPostsByUserId returns post messages by user_id
func (s *postServer) ListPostsByUserId(ctx context.Context, req *postpb.ListPostsByUserIdRequest) (*postpb.ListPostsByUserIdResponse, error) {
	userID := req.UserId

	resp, err := s.userCli.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	var postMessages []*postpb.PostMessage
	for _, up := range data.UserPosts {
		if up.UserID != userID {
			continue
		}

		for _, p := range up.Posts {
			p.Author = resp.UserMessage.Name
		}

		postMessages = up.Posts
		break
	}

	return &postpb.ListPostsByUserIdResponse{
		PostMessages: postMessages,
	}, nil
}

// ListPosts returns all post messages
func (s *postServer) ListPosts(ctx context.Context, req *postpb.ListPostsRequest) (*postpb.ListPostsResponse, error) {
	var postMessages []*postpb.PostMessage
	fmt.Println(postMessages)
	for _, up := range data.UserPosts {
		resp, err := s.userCli.GetUser(ctx, &userpb.GetUserRequest{UserId: up.UserID})
		if err != nil {
			return nil, err
		}

		for _, p := range up.Posts {
			p.Author = resp.UserMessage.Name
		}

		postMessages = append(postMessages, up.Posts...)
	}

	return &postpb.ListPostsResponse{
		PostMessages: postMessages,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":"+portNumber)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	userCli := GetUserClient("0.0.0.0:9000")
	grpcServer := grpc.NewServer()
	postpb.RegisterPostServer(grpcServer, &postServer{
		userCli: userCli,
	})

	log.Printf("start gRPC server on %s port", portNumber)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

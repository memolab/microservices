package main

import (
	"context"
	"microservices-demo/common/pb/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.messageConsumer.usersStore.m.RLock()
	defer s.messageConsumer.usersStore.m.RUnlock()

	if user, ok := s.messageConsumer.usersStore.users[req.GetId()]; ok {
		return &pb.GetUserResponse{
			User: &pb.User{
				Id:   user.ID,
				Name: user.Name,
				Data: user.Data,
			},
		}, nil
	}

	return nil, status.Error(codes.NotFound, "user not found")
}

package hitsperf

import (
	"context"
	"net"

	"github.com/QuangTung97/hits"
	"google.golang.org/grpc"
)

const (
	CommandTypeInc hits.CommandType = 1
)

type CommandInc struct {
	Value uint64
}

type Service struct {
	cmdChan chan<- hits.Command
}

func (s *Service) Inc(ctx context.Context, req *IncRequest,
) (*IncResponse, error) {
	replyChan := make(chan hits.Event, 1)

	s.cmdChan <- hits.Command{
		Type: CommandTypeInc,
		Value: CommandInc{
			Value: req.Value,
		},
		ReplyTo: replyChan,
	}

	_ = <-replyChan

	return &IncResponse{}, nil
}

func RunServer(cmdChan chan<- hits.Command) {
	s := &Service{
		cmdChan: cmdChan,
	}

	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	RegisterIncreaseServiceServer(server, s)

	err = server.Serve(listener)
	if err != nil {
		panic(err)
	}
}

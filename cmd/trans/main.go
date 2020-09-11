package main

import (
	"context"
	"hitsperf"
	"log"
	"net"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
)

type Service struct {
	db      *sqlx.DB
	counter uint64
}

func NewService() *Service {
	db := sqlx.MustConnect("mysql", "root:1@tcp(localhost:3306)/bench")
	db.SetMaxOpenConns(100)
	db.MustExec("TRUNCATE events")
	return &Service{
		db:      db,
		counter: 0,
	}
}

func (s *Service) Inc(ctx context.Context, req *hitsperf.IncRequest) (*hitsperf.IncResponse, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rollback := func() {
		atomic.AddUint64(&s.counter, ^uint64(0))
		_ = tx.Rollback()
	}

	counter := atomic.AddUint64(&s.counter, 1)
	query := `INSERT INTO events(seq, type, timestamp, data) VALUES (?, ?, ?, ?)`

	event := &hitsperf.EventIncProto{
		Value: counter,
	}

	bytes, err := proto.Marshal(event)
	if err != nil {
		rollback()
		panic(err)
	}

	timestamp := uint64(time.Now().Nanosecond() / 1000000)

	_, err = tx.ExecContext(ctx, query, counter, 1, timestamp, bytes)
	if err != nil {
		rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		rollback()
		return nil, err
	}

	log.Println("SEQ:", counter)

	return &hitsperf.IncResponse{}, nil
}

func main() {
	s := NewService()

	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	hitsperf.RegisterIncreaseServiceServer(server, s)

	err = server.Serve(listener)
	if err != nil {
		panic(err)
	}
}

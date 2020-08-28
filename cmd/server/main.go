package main

import (
	"bytes"
	"context"
	"hitsperf"
	"log"

	"github.com/QuangTung97/hits"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

type processor struct {
	counter uint64
}

func (p *processor) Init() uint64 {
	p.counter = 20

	return 0
}

const eventTypeInc hits.EventType = 1

type EventInc struct {
	Value uint64
}

func (p *processor) Process(
	cmdType hits.CommandType, cmd interface{},
	timestamp uint64,
) (eventType hits.EventType, event interface{}) {
	if cmdType != hitsperf.CommandTypeInc {
		panic("CommandType not recognized")
	}
	command := cmd.(hitsperf.CommandInc)

	p.counter += command.Value

	return eventTypeInc, EventInc{Value: command.Value}
}

func eventMarshaller(eventType hits.EventType, e interface{}) []byte {
	if eventType != eventTypeInc {
		panic("EventType not recognized")
	}
	event := e.(EventInc)

	result := &hitsperf.EventIncProto{
		Value: event.Value,
	}

	bytes, err := proto.Marshal(result)
	if err != nil {
		panic(err)
	}

	return bytes
}

type DBTest struct {
	db *sqlx.DB
}

func NewDBTest() *DBTest {
	db := sqlx.MustConnect("mysql", "root:1@tcp(localhost:3306)/bench")
	return &DBTest{
		db: db,
	}
}

func (db *DBTest) Store(events []hits.MarshalledEvent) {
	log.Println("LENGTH", len(events))

	query := `INSERT INTO events(seq, type, timestamp, data) VALUES (?, ?, ?, ?)`
	buff := bytes.NewBufferString(query)
	count := len(events)
	for i := 1; i < count; i++ {
		_, err := buff.WriteString(",(?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
	}
	query = buff.String()

	args := make([]interface{}, 0, 4*len(events))
	for _, e := range events {
		args = append(args, e.Sequence)
		args = append(args, e.Type)
		args = append(args, e.Timestamp)
		args = append(args, e.Data)
	}

	ctx := context.Background()
	_, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		panic(err)
	}
}

func (db *DBTest) ReadFrom(fromSequence uint64) ([]hits.MarshalledEvent, error) {
	return nil, nil
}

func (db *DBTest) Write(events []hits.Event) {

}

func main() {
	p := &processor{}
	db := NewDBTest()

	config := hits.Config{
		RingBufferShift: 16,
		Processor:       p,
		EventMarshaller: eventMarshaller,
		Journaler:       db,
		DBWriter:        db,
	}
	ctx := hits.NewContext(config)

	cmdChan := make(chan hits.Command, 1024)

	go ctx.Run(cmdChan)
	hitsperf.RunServer(cmdChan)
}

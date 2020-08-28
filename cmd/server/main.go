package main

import (
	"bufio"
	"hitsperf"
	"log"
	"os"

	"github.com/QuangTung97/hits"
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
	log.Println(bytes)

	return bytes
}

type DBTest struct {
	file *bufio.Writer
	f    *os.File
}

func NewDBTest() *DBTest {
	f, err := os.OpenFile("events", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	file := bufio.NewWriter(f)
	return &DBTest{
		file: file,
		f:    f,
	}
}

func (db *DBTest) Store(events []hits.MarshalledEvent) {
	for _, e := range events {
		_, err := db.file.Write(e.Data)
		if err != nil {
			panic(err)
		}
	}
	err := db.file.Flush()
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
		RingBufferShift: 10,
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

.PHONY: all

all:
	protoc --go_out=plugins=grpc,paths=source_relative:. hitsperf.proto
	go build -o server cmd/server/main.go
	go build -o client cmd/client/main.go

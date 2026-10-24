package main

import (
	"kafkaG/internal/server"
)

func main() {
	broker := server.New(":9092")
	if err := broker.Start(); err != nil {
		panic(err)
	}
}
